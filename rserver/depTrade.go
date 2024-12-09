package rserver

import (
	"context"
	"fmt"
	binance_connector "github.com/binance/binance-connector-go"
	"log"
	"muskex/gen/mproto"
	"muskex/gen/mproto/model"
	"muskex/rserver/ws"
	"muskex/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

var websocketStreamClient = binance_connector.NewWebsocketStreamClient(true)

func SubData(isDev bool) {
	binance_connector.WebsocketKeepalive = true

	errHandler := func(err error) {
		fmt.Println("ws errHandler", err)
	}
	var coinItems = map[string]int64{}
	if isDev {
		coinItems = map[string]int64{
			"BTC": 1,
			//"ETH": 2,
		}
	} else {
		coinItems = KlineIdMap
		//err := utils.Orm.Find(&coinItems, model.Coin{Status: 1}).Error
		//if err != nil {
		//	log.Println(err)
		//	return
		//}
	}
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		defer log.Println("**** stop  stop Depth  data ******")
		var symbolMapPair = map[string]string{}
		for k, _ := range coinItems {
			symbolMapPair[k+"USDT"] = "20"
		}
		for {
			log.Println("---- start  Depth  data ----")
			doneCh, stopChan, err := websocketStreamClient.WsCombinedPartialDepthServe(symbolMapPair, depEventHandle, errHandler)
			if err != nil {
				log.Println(err)
				return
			}
			if isDev {
				time.Sleep(10 * time.Second)
				stopChan <- struct{}{}
				return
			}
			<-doneCh
			log.Println("---- Depth  data stoped ----")
			time.Sleep(310 * time.Second)
		}

	}()
	//ticker data
	go func() {
		defer wg.Done()
		defer log.Println("**** stop  stop ticker  data ******")
		var symbolMapPair = []string{}
		for k, _ := range coinItems {
			symbolMapPair = append(symbolMapPair, k+"USDT")
		}
		for {
			log.Println("---- start  ticker  data ----")
			doneCh, stopChan, err := websocketStreamClient.WsCombinedMarketTickersStatServe(symbolMapPair, tickerEventHandle, errHandler)
			if err != nil {
				log.Println(err)
				return
			}
			if isDev {
				time.Sleep(20 * time.Second)
				stopChan <- struct{}{}
				return
			}
			<-doneCh
			log.Println("---- ticker  data stoped ----")
			time.Sleep(310 * time.Second)
		}
	}()
	go func() {
		defer wg.Done()
		defer log.Println("**** stop  stop AggTrade  data ******")
		var symbolMapPair = []string{}
		for k, _ := range coinItems {
			symbolMapPair = append(symbolMapPair, k+"USDT")
		}
		for {
			log.Println("---- start  AggTrade  data ----")
			doneCh, stopChan, err := websocketStreamClient.WsCombinedAggTradeServe(symbolMapPair, tradeEventHandler, errHandler)
			if err != nil {
				log.Println(err)
				return
			}
			//StartChans()
			if isDev {
				time.Sleep(30 * time.Second)
				stopChan <- struct{}{}
				return
			}
			<-doneCh
			log.Println("---- AggTrade  data stoped ----")
			time.Sleep(310 * time.Second)
		}
	}()
	wg.Wait()
	fmt.Println("******** SubData exit **********")
}

// var TradeMap = map[string][]json.RawMessage{}
var TradeMap = sync.Map{}
var TradeChan = make(chan *binance_connector.WsAggTradeEvent, 200)

func tradeEventHandler(event *binance_connector.WsAggTradeEvent) {
	key := strings.TrimSuffix(event.Symbol, "USDT")
	//item, _ := json.Marshal(event)
	//TradeMap only store 20 items, the newest item is the first item

	var items = []*binance_connector.WsAggTradeEvent{}
	v, ok := TradeMap.Load(key)
	if ok {
		items = v.([]*binance_connector.WsAggTradeEvent)
	}
	if len(items) < 20 {
		items = append([]*binance_connector.WsAggTradeEvent{event}, items...)
		//log.Println(event.Symbol, string(TradeMap[key]))
	} else {
		items = append([]*binance_connector.WsAggTradeEvent{event}, items[:19]...)
	}
	TradeMap.Store(key, items)
	ws.THub.PushMsg(event)
	//TradeChan <- event
}

var DepthMap = sync.Map{}
var DepthChan = make(chan *binance_connector.WsPartialDepthEvent, 200)

func depEventHandle(event *binance_connector.WsPartialDepthEvent) {
	key := strings.TrimSuffix(event.Symbol, "USDT")
	//bs, _ := json.Marshal(event)
	DepthMap.Store(key, event)
	//DepthChan <- event
}

var lastTickerDataTime time.Time
var TickerMap sync.Map
var TickerChan = make(chan *binance_connector.WsMarketTickerStatEvent, 200)

func tickerEventHandle(event *binance_connector.WsMarketTickerStatEvent) {
	key := strings.TrimSuffix(event.Symbol, "USDT")
	//data, _ := json.Marshal(event)
	TickerMap.Store(key, event)
	//TickerChan <- event
	lastTickerDataTime = time.Now()
}

var lastSyncCoinPriceTime time.Time

func InitKlineData() {
	coinItems := []*mproto.RankItem{}
	if err := utils.RawOrm.Find(&coinItems).Error; err != nil {
		log.Fatal(err)
	}
	for _, item := range coinItems {
		KlineIdMap[item.Name] = item.Id
		IdKlineMap[item.Id] = item.Name
	}
}

var KlineIdMap = map[string]int64{}
var IdKlineMap = map[int64]string{}

func CoinIdMap() {
	//var coinIdMap = map[string]int64{}
	//coinItems1 := []*model.Coin{}
	//if err := utils.RawOrm.Find(&coinItems1).Error; err != nil {
	//	log.Fatal(err)
	//}
	//for _, item := range coinItems1 {
	//	CoinIdMap[item.Name] = item.Id
	//	IdCoinMap[item.Id] = item.Name
	//}
}
func IdCoinMap(ctx context.Context) map[int64]string {
	res, _ := utils.CacheFromLruWithFixKey("IdCoinMap", func() (interface{}, error) {
		var idCoinMap = map[int64]string{}
		//var coinIdMap = map[string]int64{}
		coinItems1 := []*model.Coin{}
		if err := utils.Orm.Find(&coinItems1).Error; err != nil {
			log.Fatal(err)
		}
		for _, item := range coinItems1 {
			//CoinIdMap[item.Name] = item.Id
			idCoinMap[item.Id] = item.Name
		}
		return idCoinMap, nil
	})
	return res.(map[int64]string)
}

func SyncCoinPrice() {
	for {
		if lastSyncCoinPriceTime.Before(lastTickerDataTime) {
			log.Println("syncCoinPrice")
			list := []*tickerCoin{}
			TickerMap.Range(func(key, value interface{}) bool {
				//bs := value.([]byte)
				//ticker := &binance_connector.WsMarketTickerStatEvent{}
				//err := json.Unmarshal(bs, ticker)
				//if err != nil {
				//	log.Fatal(err)
				//}
				ticker := value.(*binance_connector.WsMarketTickerStatEvent)
				id := KlineIdMap[key.(string)]
				if id > 0 {
					price, _ := strconv.ParseFloat(ticker.LastPrice, 64)
					change, _ := strconv.ParseFloat(ticker.PriceChangePercent, 64)
					total, _ := strconv.ParseFloat(ticker.QuoteVolume, 64)
					list = append(list, &tickerCoin{
						Id: id, InitialPrice: price,
						EventTime:          uint64(ticker.Time / 1000),
						PriceChangePercent: float32(change),
						TotalTrade:         total,
					})
				}
				return true
			})
			if len(list) == 0 {
				log.Println("no coins price updated")
				time.Sleep(2 * time.Second)
				continue
			}
			dbtx := utils.RawOrm.Begin()
			res := dbtx.Save(&list)
			if res.Error != nil {
				log.Println(res.Error)
				dbtx.Rollback()
			}
			dbtx.Commit()
			lastSyncCoinPriceTime = time.Now()
			log.Println(len(list), "coins price updated", res.RowsAffected)
		} else {
			log.Println("no new ticker data,skip SyncCoinPrice")
			time.Sleep(3 * time.Second)
		}
		time.Sleep(2 * time.Second)
	}
}
func (*tickerCoin) TableName() string {
	return "ba_rank_item"
}

type tickerCoin struct {
	Id                 int64
	InitialPrice       float64
	EventTime          uint64
	PriceChangePercent float32
	TotalTrade         float64
}
