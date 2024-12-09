package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
)

type Data struct {
	EventType             string  `json:"e"`
	EventTime             int64   `json:"E"`
	Symbol                string  `json:"s"`
	PriceChange           string  `json:"p"`
	PriceChangePercent    string  `json:"P"`
	WeightedAvgPrice      string  `json:"w"`
	FirstTradePrice       string  `json:"x"`
	LastPrice             string  `json:"c"`
	LastQty               string  `json:"Q"`
	BestBidPrice          string  `json:"b"`
	BestBidQty            string  `json:"B"`
	BestAskPrice          string  `json:"a"`
	BestAskQty            string  `json:"A"`
	OpenPrice             string  `json:"o"`
	HighPrice             string  `json:"h"`
	LowPrice              string  `json:"l"`
	TotalTradedBaseAsset  string  `json:"v"`
	TotalTradedQuoteAsset string  `json:"q"`
	StatisticsOpenTime    int64   `json:"O"`
	StatisticsCloseTime   int64   `json:"C"`
	FirstTradeID          int64   `json:"F"`
	LastTradeID           int64   `json:"L"`
	TradeCount            int64   `json:"n"`
	Usd                   float64 `json:"usd"`
}

type Ticker struct {
	Stream string `json:"stream"`
	Data   Data   `json:"data"`
}

type OutTicker struct {
	KlineType string `json:"kline_type"`
	Data      Data   `json:"data"`
	ID        int    `json:"id"`
	LogoImage string `json:"logo_image"`
	Alias     string `json:"alias"`
}

type Response struct {
	Code int                    `json:"code"`
	Data map[string][]OutTicker `json:"data"`
	Msg  string                 `json:"msg"`
	Page string                 `json:"page"`
	Time int64                  `json:"time"`
}

type Item struct {
	ID        int    `json:"id"`
	LogoImage string `json:"logo_image"`
	KlineType string `json:"kline_type"`
	Margin    int    `json:"margin"`
	Data      Data   `json:"data"`
	Alias     string `json:"alias"`
}

type ItemNoData struct {
	ID        int    `json:"id"`
	LogoImage string `json:"logo_image"`
	KlineType string `json:"kline_type"`
	Margin    int    `json:"margin"`
	Alias     string `json:"alias"`
}

type ResponseTickter struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Page string `json:"page"`
	Time int64  `json:"time"`
	Data []Item `json:"data"`
}

type ByMultipleFields struct {
	Coins     []ItemNoData
	AscMargin bool
	AscName   bool
}

func (b ByMultipleFields) Len() int      { return len(b.Coins) }
func (b ByMultipleFields) Swap(i, j int) { b.Coins[i], b.Coins[j] = b.Coins[j], b.Coins[i] }
func (b ByMultipleFields) Less(i, j int) bool {
	if b.Coins[i].Margin != b.Coins[j].Margin {
		if b.AscMargin {
			return b.Coins[i].Margin < b.Coins[j].Margin
		}
		return b.Coins[i].Margin > b.Coins[j].Margin
	}
	return false
}

var addr = "13.229.187.76:56379"
var password = "YfJJExSdM"
var ctx = context.Background()
var rdb = redis.NewClient(&redis.Options{
	Addr:     addr,
	Password: password,
	DB:       10,
})

var redisClient *redis.Client

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       12,
	})
	//logFile, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//if err != nil {
	//	log.Fatal("Failed to open log file:", err)
	//}
	//log.SetOutput(logFile)
}

type Subscription struct {
	conn    *websocket.Conn
	send    chan []byte
	topics  map[string]bool
	once    sync.Once
	closeCh chan struct{}
}

func (sub *Subscription) close() {
	defer func() {
		if r := recover(); r != nil {
			// 捕获关闭已关闭的通道导致的 panic
			fmt.Println("Recovered from panic in cancelSubscription:", r)
		}
	}()
	sub.once.Do(func() {
		if sub.closeCh != nil {
			select {
			case <-sub.closeCh:
				log.Println("通道已经关闭")
			default:
				close(sub.closeCh)
				log.Println("通道关闭成功")
			}
		}
		close(sub.send)
		sub.conn.Close()
	})
}

type Hub struct {
	subscriptions map[string][]*Subscription
	mu            sync.RWMutex
}

func newHub() *Hub {
	return &Hub{
		subscriptions: make(map[string][]*Subscription),
	}
}
func (h *Hub) subscribe(topic string, sub *Subscription) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, exists := sub.topics[topic]; exists {
		log.Printf("已经订阅主题: %s", topic)
		return
	}
	h.subscriptions[topic] = append(h.subscriptions[topic], sub)
	sub.topics[topic] = true
	log.Printf("客户端订阅主题: %s", topic)
	if topic == "ticker" {
		go sub.ticker()
	}
	if topic == "trading" {
		go sub.trading()
	}
	if topic == "increase" {
		go sub.increase()
	}
	if topic == "decrease" {
		go sub.decrease()
	}
}
func (h *Hub) unsubscribe(topic string, sub *Subscription) {
	h.mu.Lock()
	subs := h.subscriptions[topic]
	for i, s := range subs {
		if s == sub {
			h.subscriptions[topic] = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	delete(sub.topics, topic)
	log.Printf("客户端取消订阅主题: %s", topic)
	h.mu.Unlock()
	//sub.close() // 解锁后关闭订阅者，避免潜在死锁
	sub.cancelSubscription()
}

func (sub *Subscription) cancelSubscription() {
	defer func() {
		if r := recover(); r != nil {
			// 捕获关闭已关闭的通道导致的 panic
			fmt.Println("Recovered from panic in cancelSubscription:", r)
		}
	}()

	// 确保通道只关闭一次
	if sub.closeCh != nil {
		close(sub.closeCh)
		sub.closeCh = nil
	}
}

func (h *Hub) publish(topic string, message []byte) {
	h.mu.RLock()
	subs, ok := h.subscriptions[topic]
	h.mu.RUnlock()
	if !ok {
		return
	}
	// 创建订阅者的副本，避免遍历时发生并发问题
	subsCopy := make([]*Subscription, len(subs))
	copy(subsCopy, subs)
	for _, sub := range subsCopy {
		select {
		case sub.send <- message:
			// 成功发送消息
			log.Printf("消息已发送到主题 %s", topic)
		case <-sub.closeCh:
			// 订阅者已关闭
			log.Printf("订阅通道已关闭，主题 %s", topic)
			h.unsubscribe(topic, sub) // 移除订阅者
		default:
			// 发送失败，可能通道已满或已关闭
			log.Printf("无法发送消息到主题 %s，可能通道已关闭", topic)
			h.unsubscribe(topic, sub) // 移除订阅者
		}
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main1() {
	hub := newHub()
	go func() {
		for {
			time.Sleep(1 * time.Second)
			randomData := fetchDataFromRedis()
			hub.publish("home", randomData)
		}
	}()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		fmt.Println("token:", token)
		if !isValidToken(token) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Upgrade:", err)
			return
		}
		sub := &Subscription{
			conn:   conn,
			send:   make(chan []byte, 256),
			topics: make(map[string]bool),
		}
		go handleMessages(hub, sub)
		go sendMessages(sub)
		//go sendHeartbeat(sub)
	})
	log.Println("Server started at :8082")
	log.Fatal(http.ListenAndServe("0.0.0.0:8082", nil))
}
func isValidToken(token string) bool {
	if token == "" {
		return false
	}
	_, err := redisClient.Get(ctx, "token:"+token).Result()
	if err != nil {
		if err == redis.Nil {
			return false // Token 不存在
		}
		log.Println("Redis Get error:", err)
		return false
	}
	return true
}
func handleMessages(hub *Hub, sub *Subscription) {
	defer func() {
		log.Println("Closing connection")
		sub.close()
	}()
	for {
		_, message, err := sub.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("Connection closed:", err)
			} else {
				log.Println("Read message error:", err)
			}
			break
		}
		var msg map[string]string
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Parse message error:", err)
			continue
		}
		if cmd, ok := msg["sub"]; ok {
			hub.subscribe(cmd, sub)
		} else if cmd, ok := msg["cancel"]; ok {
			hub.unsubscribe(cmd, sub)
		}
	}
}
func sendMessages(sub *Subscription) {
	for msg := range sub.send {
		if msg == nil {
			log.Println("Send channel closed, stopping message sending")
			return
		}
		err := sub.conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("发送消息错误:", err)
			break
		}
	}
}
func sendHeartbeat(sub *Subscription) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if sub.conn == nil {
				log.Println("WebSocket 连接为 nil，无法发送心跳")
				return
			}

			response := map[string]string{
				"type":      "heartbeat",
				"timestamp": time.Now().Format(time.RFC3339),
			}
			responseMessage, err := json.Marshal(response)
			if err != nil {
				log.Println("序列化心跳消息错误:", err)
				continue
			}

			err = sub.conn.WriteMessage(websocket.TextMessage, responseMessage)
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Println("连接关闭:", err)
				} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					log.Println("写入超时:", err)
				} else {
					log.Println("WriteMessage 错误:", err)
				}
				sub.conn.Close()
				return
			}
		}
	}
}
func (sub *Subscription) ticker() {
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
		sub.close()
	}()
	for {
		select {
		case <-ticker.C:
			if sub.topics["ticker"] {
				log.Println("Publishing ticker")
				coinList := getCoinList()
				if coinList != nil {
					quoteData := fetchTickerFromRedis(coinList)
					select {
					case sub.send <- quoteData:
						log.Println("Quote data sent")
					case <-sub.closeCh:
						log.Println("Subscription closed, stopping publishing")
						return
					default:
						log.Println("Failed to send quote data, channel might be closed")
						return
					}
				}
			}
		case <-sub.closeCh:
			log.Println("Stopping quote publishing due to subscription close")
			return
		}
	}
}
func (sub *Subscription) trading() {
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
		sub.close()
	}()
	for {
		select {
		case <-ticker.C:
			if sub.topics["trading"] {
				log.Println("Publishing trading")
				coinList := getCoinList()
				if coinList != nil {
					quoteData := fetchTradingFromRedis(coinList)
					select {
					case sub.send <- quoteData:
						log.Println("Quote data sent")
					case <-sub.closeCh:
						log.Println("Subscription closed, stopping publishing")
						return
					default:
						log.Println("Failed to send quote data, channel might be closed")
						return
					}
				}
			}
		case <-sub.closeCh:
			log.Println("Stopping quote publishing due to subscription close")
			return
		}
	}
}
func (sub *Subscription) increase() {
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
		sub.close()
	}()
	for {
		select {
		case <-ticker.C:
			if sub.topics["increase"] {
				log.Println("Publishing increase")
				coinList := getCoinList()
				if coinList != nil {
					quoteData := fetchIncreaseFromRedis(coinList)
					select {
					case sub.send <- quoteData:
						log.Println("Quote data sent")
					case <-sub.closeCh:
						log.Println("Subscription closed, stopping publishing")
						return
					default:
						log.Println("Failed to send quote data, channel might be closed")
						return
					}
				}
			}
		case <-sub.closeCh:
			log.Println("Stopping quote publishing due to subscription close")
			return
		}
	}
}
func (sub *Subscription) decrease() {
	ticker := time.NewTicker(1 * time.Second)
	defer func() {
		ticker.Stop()
		sub.close()
	}()
	for {
		select {
		case <-ticker.C:
			if sub.topics["decrease"] {
				log.Println("Publishing decrease")
				coinList := getCoinList()
				if coinList != nil {
					quoteData := fetchDecreaseFromRedis(coinList)
					select {
					case sub.send <- quoteData:
						log.Println("Quote data sent")
					case <-sub.closeCh:
						log.Println("Subscription closed, stopping publishing")
						return
					default:
						log.Println("Failed to send quote data, channel might be closed")
						return
					}
				}
			}
		case <-sub.closeCh:
			log.Println("Stopping quote publishing due to subscription close")
			return
		}
	}
}
func getCoinList() []ItemNoData {
	coinList, err := rdb.Get(ctx, "coinList").Result()
	if err != nil {
		if err == redis.Nil {
			log.Println("key does not exist")
		} else {
			log.Println("error fetching from Redis: ", err)
		}
		return nil
	} else {
		var itemNoData []ItemNoData
		err := json.Unmarshal([]byte(coinList), &itemNoData)
		if err != nil {
			fmt.Println("Error unmarshaling JSON:", err)
			return nil
		}
		return itemNoData
	}
}
func gwtUsdt2UsdRate() float64 {
	usdtToUsdRateStr, err := rdb.Get(ctx, "usdt_to_usd_rate").Result()
	if err != nil {
		if err == redis.Nil {
			log.Println("key does not exist")
		} else {
			log.Println("error fetching from Redis: ", err)
		}
		return 0
	}
	usdtToUsdRate, err := strconv.ParseFloat(usdtToUsdRateStr, 64)
	if err != nil {
		log.Println("error converting string to float64: ", err)
		return 0
	}
	return usdtToUsdRate
}

func getHomeKlineTypes() []string {
	return []string{"BTC/USDT", "ETH/USDT", "BNB/USDT"}
}
func fetchDataFromRedis() []byte {
	result, err := rdb.Get(ctx, "ticker").Result()
	if err != nil {
		if err == redis.Nil {
			log.Println("key does not exist")
		} else {
			log.Println("error fetching from Redis: ", err)
		}
		return []byte("")
	} else {
		var Maptickers map[string]Ticker
		err := json.Unmarshal([]byte(result), &Maptickers)
		if err != nil {
			fmt.Println("Error:", err)
			return []byte("")
		}
		itemNoData := getCoinList()
		var itemMap = make(map[string]ItemNoData)
		for _, item := range itemNoData {
			itemMap[item.KlineType] = item
		}
		rn := make(map[string][]OutTicker)
		var ticker OutTicker
		usdt_to_usd_rate := gwtUsdt2UsdRate()
		home_kline_types := getHomeKlineTypes()
		for index, value := range home_kline_types {
			key := strings.ReplaceAll(value, "/", "")
			key = strings.ToLower(key)
			key = key + "@ticker"
			v, exists := Maptickers[key]
			if exists {
				fmt.Printf("Key '%s' exists with value %v\n", key, v)
				ticker.Data = v.Data
				LastPrice, err := strconv.ParseFloat(v.Data.LastPrice, 64)
				if err != nil {
					log.Println("error converting string to float64: ", err)
					return []byte("")
				}
				ticker.Data.Usd = LastPrice * usdt_to_usd_rate
				ticker.KlineType = value
				itemNoData, exist := itemMap[value]
				if exist {
					ticker.ID = itemNoData.ID
					ticker.LogoImage = itemNoData.LogoImage
					ticker.Alias = itemNoData.Alias
				}
				rn["homeTicker"] = append(rn["homeTicker"], ticker)
			} else {
				fmt.Printf("Key '%s' does not exist\n", key)
			}
			fmt.Printf("Index %v: %s\n", index, value)
		}
		var response Response
		response.Data = rn
		response.Msg = "success"
		response.Time = time.Now().Unix()
		response.Code = 1
		response.Page = "home"
		jsonData, err := json.Marshal(response)
		if err != nil {
			log.Println("Error marshalling data:", err)
		}
		return jsonData
	}
}
func getSymbolMap() map[string]int {
	return map[string]int{
		"CKBUSDT":    1,
		"ONEUSDT":    1,
		"ZILUSDT":    1,
		"JASMYUSDT":  1,
		"BNBUSDT":    1,
		"ETHUSDT":    1,
		"BTCUSDT":    1,
		"SOLUSDT":    1,
		"XRPUSDT":    1,
		"DOGEUSDT":   1,
		"LINKUSDT":   1,
		"ADAUSDT":    1,
		"SHIBUSDT":   1,
		"AVAXUSDT":   1,
		"WBTCUSDT":   1,
		"BCHUSDT":    1,
		"NEARUSDT":   1,
		"MATICUSDT":  1,
		"LTCUSDT":    1,
		"ICPUSDT":    1,
		"PEPEUSDT":   1,
		"UNIUSDT":    1,
		"RNDRUSDT":   1,
		"ETCUSDT":    1,
		"HBARUSDT":   1,
		"ATOMUSDT":   1,
		"APTUSDT":    1,
		"FILUSDT":    1,
		"XLMUSDT":    1,
		"IMXUSDT":    1,
		"STXUSDT":    1,
		"MKRUSDT":    1,
		"GRTUSDT":    1,
		"OPUSDT":     1,
		"VETUSDT":    1,
		"TAOUSDT":    1,
		"INJUSDT":    1,
		"THETAUSDT":  1,
		"FTMUSDT":    1,
		"RUNEUSDT":   1,
		"BONKUSDT":   1,
		"TIAUSDT":    1,
		"LDOUSDT":    1,
		"ALGOUSDT":   1,
		"FLOWUSDT":   1,
		"ARBUSDT":    1,
		"AAVEUSDT":   1,
		"GALAUSDT":   1,
		"SEIUSDT":    1,
		"SUIUSDT":    1,
		"ENAUSDT":    1,
		"NEOUSDT":    1,
		"CHZUSDT":    1,
		"EGLDUSDT":   1,
		"PEOPLEUSDT": 1,
		"WIFUSDT":    1,
		"WLDUSDT":    1,
		"TRBUSDT":    1,
		"BOMEUSDT":   1,
		"ORDIUSDT":   1,
		"JTOUSDT":    1,
		"DOTUSDT":    1,
		"ARUSDT":     1,
		"QNTUSDT":    1,
	}
}
func fetchTicker(itemNoData []ItemNoData) []Item {
	// 排序条件
	ascendingMargin := false // SORT_DESC
	ascendingName := true    // SORT_ASC
	// 使用 sort.Sort 进行排序
	sort.Sort(ByMultipleFields{Coins: itemNoData, AscMargin: ascendingMargin, AscName: ascendingName})
	result, err := rdb.Get(ctx, "ticker").Result()
	if err != nil {
		if err == redis.Nil {
			log.Println("key does not exist")
		} else {
			log.Println("error fetching from Redis: ", err)
		}
		return nil
	} else {
		var Maptickers map[string]Ticker
		err := json.Unmarshal([]byte(result), &Maptickers)
		if err != nil {
			fmt.Println("Error:", err)
			return nil
		}
		symbolMap := getSymbolMap()
		rn := make([]Item, 0)
		var item Item
		usdt_to_usd_rate := gwtUsdt2UsdRate()
		for index, value := range itemNoData {
			key := strings.ReplaceAll(value.KlineType, "/", "")
			key = strings.ToLower(key)
			Bkey := strings.ToUpper(key)
			_, ex := symbolMap[Bkey]
			if ex {
				key = key + "@ticker"
				v, exists := Maptickers[key]
				if exists {
					fmt.Printf("Key '%s' exists with value %v\n", key, v)
					item.Data = v.Data
					LastPrice, err := strconv.ParseFloat(v.Data.LastPrice, 64)
					if err != nil {
						log.Println("error converting string to float64: ", err)
						return nil
					}
					item.Data.Usd = LastPrice * usdt_to_usd_rate
					item.KlineType = value.KlineType
					item.LogoImage = value.LogoImage
					item.Margin = value.Margin
					item.ID = value.ID
					item.Alias = value.Alias
					rn = append(rn, item)
				} else {
					fmt.Printf("Key '%s' does not exist\n", key)
				}
				fmt.Printf("Index %d: %v\n", index, value)
			}
		}
		return rn
	}
}
func fetchTickerFromRedis(itemNoData []ItemNoData) []byte {
	rn := fetchTicker(itemNoData)
	sort.Slice(rn, func(i, j int) bool {
		qi, _ := strconv.ParseFloat(rn[i].Data.LastPrice, 64)
		qj, _ := strconv.ParseFloat(rn[j].Data.LastPrice, 64)
		return qi > qj // 降序
	})
	var response ResponseTickter
	response.Data = rn
	response.Msg = "success"
	response.Time = time.Now().Unix()
	response.Code = 1
	response.Page = "ticker"
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Println("Error marshalling data:", err)
	}
	return jsonData
}
func get8Item(itemNoData []Item) []Item {
	var firstEight []Item
	if len(itemNoData) > 8 {
		firstEight = itemNoData[:8]
	} else {
		firstEight = itemNoData
	}
	return firstEight
}
func fetchTradingFromRedis(itemNoData []ItemNoData) []byte {
	rn := fetchTicker(itemNoData)
	sort.Slice(rn, func(i, j int) bool {
		qi, _ := strconv.ParseFloat(rn[i].Data.TotalTradedQuoteAsset, 64)
		qj, _ := strconv.ParseFloat(rn[j].Data.TotalTradedQuoteAsset, 64)
		return qi > qj // 降序
	})
	firstEight := get8Item(rn)
	var response ResponseTickter
	response.Data = firstEight
	response.Msg = "success"
	response.Time = time.Now().Unix()
	response.Code = 1
	response.Page = "trading"
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Println("Error marshalling data:", err)
	}
	return jsonData
}
func fetchIncreaseFromRedis(itemNoData []ItemNoData) []byte {
	rn := fetchTicker(itemNoData)
	sort.Slice(rn, func(i, j int) bool {
		qi, _ := strconv.ParseFloat(rn[i].Data.PriceChangePercent, 64)
		qj, _ := strconv.ParseFloat(rn[j].Data.PriceChangePercent, 64)
		return qi > qj // 降序
	})
	firstEight := get8Item(rn)
	var response ResponseTickter
	response.Data = firstEight
	response.Msg = "success"
	response.Time = time.Now().Unix()
	response.Code = 1
	response.Page = "increase"
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Println("Error marshalling data:", err)
	}
	return jsonData
}
func fetchDecreaseFromRedis(itemNoData []ItemNoData) []byte {
	rn := fetchTicker(itemNoData)
	sort.Slice(rn, func(i, j int) bool {
		qi, _ := strconv.ParseFloat(rn[i].Data.PriceChangePercent, 64)
		qj, _ := strconv.ParseFloat(rn[j].Data.PriceChangePercent, 64)
		return qi < qj // 升序
	})
	firstEight := get8Item(rn)
	var response ResponseTickter
	response.Data = firstEight
	response.Msg = "success"
	response.Time = time.Now().Unix()
	response.Code = 1
	response.Page = "decrease"
	jsonData, err := json.Marshal(response)
	if err != nil {
		log.Println("Error marshalling data:", err)
	}
	return jsonData
}
