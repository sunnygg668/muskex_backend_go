package main

import (
	"fmt"
	binance_connector "github.com/binance/binance-connector-go"
	"github.com/spf13/pflag"
	"log"
	"muskex/gen/mproto"
	"muskex/utils"
	"os"
	"strconv"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)
	printFlags := false
	var dbUrl, serverPort string
	var defaultPdbs = []string{
		"admin:UhwZR0BApOSD57qATCao@tcp(musk-ex2-2024-11-155.cuyych8yxu1j.ap-southeast-1.rds.amazonaws.com:3306)/musk_test?loc=Local&parseTime=true&multiStatements=true",
		"admin:UhwZR0BApOSD57qATCao@tcp(musk-ex2-2024-11-155.cuyych8yxu1j.ap-southeast-1.rds.amazonaws.com:3306)/btlux?loc=Local&parseTime=true&multiStatements=true",
		"admin:UhwZR0BApOSD57qATCao@tcp(musk-ex2-2024-11-155.cuyych8yxu1j.ap-southeast-1.rds.amazonaws.com:3306)/bzone?loc=Local&parseTime=true&multiStatements=true",
		"admin:UhwZR0BApOSD57qATCao@tcp(musk-ex2-2024-11-155.cuyych8yxu1j.ap-southeast-1.rds.amazonaws.com:3306)/axex?loc=Local&parseTime=true&multiStatements=true",
		"admin:UhwZR0BApOSD57qATCao@tcp(musk-ex2-2024-11-155.cuyych8yxu1j.ap-southeast-1.rds.amazonaws.com:3306)/ascend_ex?loc=Local&parseTime=true&multiStatements=true",
	}
	//--db 'admin:BMkOfAnChSqsC84BgdLJ@tcp(one-database-1.cvqm0oyo6ebo.ap-southeast-1.rds.amazonaws.com:3306)/kline?loc=Local&parseTime=true&multiStatements=true'  test2 -L 3306:one-database-1.cvqm0oyo6ebo.ap-southeast-1.rds.amazonaws.com:3306
	//pflag.StringVarP(&dbUrl, "db", "d", "admin:BMkOfAnChSqsC84BgdLJ@tcp(localhost:3306)/kline?loc=Local&parseTime=true&multiStatements=true", "mysql database url")
	pflag.StringVarP(&dbUrl, "db", "d", "admin:UhwZR0BApOSD57qATCao@tcp(musk-ex2-2024-11-155.cuyych8yxu1j.ap-southeast-1.rds.amazonaws.com:3306)/saas?loc=Local&parseTime=true&multiStatements=true", "mysql database url")
	pflag.StringArrayVarP(&pdbs, "pdbs", "", defaultPdbs, "")
	pflag.StringVarP(&serverPort, "port", "p", "8080", "api　service port")
	pflag.BoolVarP(&printFlags, "print", "", false, "")
	pflag.Parse()
	if printFlags {
		log.Println("db:", dbUrl)
		log.Println("pdbs:")
		for _, pdb := range pdbs {
			log.Println("db:", pdb)
		}
		os.Exit(0)
	}
	utils.InitDb(dbUrl)
	//utils.Orm.AutoMigrate(KlineData{})
	websocketStreamClient = binance_connector.NewWebsocketStreamClient(true)
	binance_connector.WebsocketKeepalive = true
}

var websocketStreamClient *binance_connector.WebsocketStreamClient

var pdbs []string

func main() {
	errHandler := func(err error) {
		fmt.Println(err)
	}
	var coinItems = []mproto.RankItem{}
	utils.Orm.Find(&coinItems)
	AggCons(pdbs)
	go tableClean(pdbs)
	go batchWrite()
	go func() {
		var symbolMapPair = map[string]string{}
		for _, item := range coinItems {
			symbolMapPair[item.Name+"USDT"] = "1m"
		}
		doneCh, _, err := websocketStreamClient.WsCombinedKlineServe(symbolMapPair, dataHandler, errHandler)
		if err != nil {
			fmt.Println(err)
			return
		}
		<-doneCh
		fmt.Println("**** stop  stop kline 1m data ******")
	}()

	var symbolMapPair = map[string]string{}
	for _, item := range coinItems {
		symbolMapPair[item.Name+"USDT"] = "1M"
	}
	doneCh, _, err := websocketStreamClient.WsCombinedKlineServe(symbolMapPair, dataHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	<-doneCh
	fmt.Println("**** stop  stop kline 1M data ******")

	//go func() {
	//	time.Sleep(14 * time.Second)
	//	stopCh <- struct{}{}
	//}()
}
func tableClean(pdbs []string) {
	sql := `delete
from ba_kline_data
where id in (select aa.id
             from (select row_number() over (partition by symbol,stat_type order by start_time desc) rownum, t.*
                   from ba_kline_data t) aa
             where aa.rownum > 500);
`
	//run sql every 10 minite
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			res := utils.Orm.Exec(sql)
			AggCons(pdbs)
			if res.Error != nil {
				log.Println("tableClean", res.Error)
			}
			log.Println("tableClean RowsAffected", res.RowsAffected)
		}
	}

}

func getStreamKey(symbol string, interval string) string {
	return symbol + "@kline_" + interval
}

var klineDataChan = make(chan *KlineData, 200)
var dataHandler = func(event *binance_connector.WsKlineEvent) {
	//fmt.Println(binance_connector.PrettyPrint(event))
	open, _ := strconv.ParseFloat(event.Kline.Open, 64)
	close, _ := strconv.ParseFloat(event.Kline.Close, 64)
	high, _ := strconv.ParseFloat(event.Kline.High, 64)
	low, _ := strconv.ParseFloat(event.Kline.Low, 64)
	volume, _ := strconv.ParseFloat(event.Kline.Volume, 64)
	quoteVolume, _ := strconv.ParseFloat(event.Kline.QuoteVolume, 64)

	klog := KlineData{
		StartTime:   event.Kline.StartTime / 1000,
		Symbol:      event.Symbol[:len(event.Symbol)-4],
		Open:        open,
		Close:       close,
		High:        high,
		Low:         low,
		Volume:      volume,
		StatType:    event.Kline.Interval,
		QuoteVolume: quoteVolume,
	}
	klineDataChan <- &klog

}

// write to db, batch write every 1s
func batchWrite() {
	item_datas := make([]*KlineData, 0, 200)
	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			res2 := utils.Orm.Save(item_datas)
			if res2.Error != nil {
				log.Println("CreateInBatches datas", res2.Error)
			}
			log.Println("batchWrite datas count", len(item_datas))
			item_datas = make([]*KlineData, 0, 200)
			batchWrite1()
		case log := <-klineDataChan:
			//fmt.Println(log)
			item_datas = append(item_datas, log)
		}
	}
	log.Println("batchWrite exit")
}
func batchWrite1() {
	//res := utils.Orm.Exec(sql1)
	//if res.Error != nil {
	//	log.Println("sql1", res.Error)
	//}
	//log.Println("sql1 RowsAffected", res.RowsAffected)

	for _, param := range params_sql2 {
		if param.Name == "1w" {
			res2 := utils.Orm.Exec(sql2_week, param)
			if res2.Error != nil {
				log.Println("sql2_week", res2.Error)
			}
			log.Println("sql2_week RowsAffected", param.Name, res2.RowsAffected)
			continue
		}
		res2 := utils.Orm.Exec(sql2, param)
		if res2.Error != nil {
			log.Println("sql2", res2.Error)
		}
		log.Println("sql2 RowsAffected", param.Name, res2.RowsAffected)
	}

}

//	type KlineLog struct {
//		Id        int64 `json:"id"`
//		StartTime int64 `json:"t"`
//		//EndTime              int64  `json:"T"`
//		Symbol string `json:"s"`
//		//Interval             string `json:"i"`
//		//FirstTradeID         int64  `json:"f"`
//		//LastTradeID          int64  `json:"L"`
//		Open  float64 `json:"o"`
//		Close float64 `json:"c"`
//		High  float64 `json:"h"`
//		Low   float64 `json:"l"`
//		//成交量
//		Volume float64 `json:"v"`
//		//TradeNum             int64  `json:"n"`
//
//		IsFinal bool `json:"x"`
//		//成交额
//		QuoteVolume float64 `json:"q"`
//		//ActiveBuyVolume      string `json:"V"`
//		//ActiveBuyQuoteVolume string `json:"Q"`
//	}
type KlineData struct {
	Id int64 `json:"id"`
	//'1m', '5m', '15m', '30m', '1h', '4h', '1d'
	StatType string `json:"i"`
	Symbol   string `json:"s"`
	//unix timestamp ; Interval startDate
	StartTime int64 `json:"t"`

	Open  float64 `json:"o"`
	Close float64 `json:"c"`
	High  float64 `json:"h"`
	Low   float64 `json:"l"`
	//成交量
	Volume float64 `json:"v"`
	//成交额
	QuoteVolume float64 `json:"q"`
}

var sql1 = `
insert into ba_kline_data(stat_type, symbol, start_time, open, close, high, low, volume, quote_volume)
select *
from (select stat_type,
             symbol,
             start_time,
             open1                                     open,
             close1                                    close,
             max(high)                                 high,
             min(low)                                  low,
             cast(sum(volume) as decimal(16, 2))       volume,
             cast(sum(quote_volume) as decimal(16, 2)) quote_volume
      from (select id,
                   '1m' as                  stat_type,
                   symbol,
                   volume,
                   quote_volume,
                   start_time,
                   open,
                   close,
                   first_value(open) over w open1,
                   last_value(close) over w close1,
                   high,
                   low
            from ba_kline_log
            where start_time > truncate(unix_timestamp() / 60, 0) * 60 - 60
            WINDOW w AS ( partition by symbol,start_time order by id
                    rows between unbounded preceding and unbounded following )) t
      group by symbol, start_time) abc
ON DUPLICATE KEY UPDATE close=abc.close,
                        high=abc.high,
                        low=abc.low,
                        volume=abc.volume,
                        quote_volume=abc.quote_volume;
`
var sql2 = `
insert into ba_kline_data(stat_type, symbol, start_time, open, close, high, low, volume, quote_volume)
select *
from (select stat_type,
             symbol,
             start_time,
             open1                                     open,
             close1                                    close,
             max(high)                                 high,
             min(low)                                  low,
             cast(sum(volume) as decimal(16, 2))       volume,
             cast(sum(quote_volume) as decimal(16, 2)) quote_volume
      from (select stat_type,
                   symbol,
                   start_time,
                   open,
                   close,
                   first_value(open) over w open1,
                   last_value(close) over w close1,
                   high                     high,
                   low                      low,
                   volume                   volume,
                   quote_volume             quote_volume
            from (select id,
                         @Name as                             stat_type,
                         truncate(start_time / @Value, 0) * @Value start_time,
                         symbol,
                         volume,
                         quote_volume,
                         open,
                         close,
                         high,
                         low
                  from ba_kline_data
                  where start_time >= truncate(unix_timestamp() / @Value, 0) * @Value - @Value
                    and stat_type = @PreName) aa
            WINDOW w AS ( partition by symbol,start_time order by id
                    rows between unbounded preceding and unbounded following )) t
      group by symbol, start_time) abc
ON DUPLICATE KEY UPDATE close=abc.close,
                        high=abc.high,
                        low=abc.low,
                        volume=abc.volume,
                        quote_volume=abc.quote_volume;
`
var sql2_week = `
insert into ba_kline_data(stat_type, symbol, start_time, open, close, high, low, volume, quote_volume)
select *
from (select stat_type,
             symbol,
             start_time,
             open1                                     open,
             close1                                    close,
             max(high)                                 high,
             min(low)                                  low,
             cast(sum(volume) as decimal(16, 2))       volume,
             cast(sum(quote_volume) as decimal(16, 2)) quote_volume
      from (select stat_type,
                   symbol,
                   start_time,
                   open,
                   close,
                   first_value(open) over w open1,
                   last_value(close) over w close1,
                   high                     high,
                   low                      low,
                   volume                   volume,
                   quote_volume             quote_volume
            from (select id,
                         @Name as                             stat_type,
                         truncate(start_time / @Value, 0) * @Value - 86400*3 start_time,
                         symbol,
                         volume,
                         quote_volume,
                         open,
                         close,
                         high,
                         low
                  from ba_kline_data
                  where start_time >= truncate(unix_timestamp() / @Value, 0) * @Value -86400*3 - @Value
                    and stat_type = @PreName) aa
            WINDOW w AS ( partition by symbol,start_time order by id
                    rows between unbounded preceding and unbounded following )) t
      group by symbol, start_time) abc
ON DUPLICATE KEY UPDATE close=abc.close,
                        high=abc.high,
                        low=abc.low,
                        volume=abc.volume,
                        quote_volume=abc.quote_volume;
`

// # '1m', '5m', '15m', '30m', '1h', '4h', '1d'   1w    1M
// #  60   300   900   1800    3600  14400  86400 604800
type namedParam struct {
	PreName string
	Name    string
	Value   int
}

var params_sql2 = []namedParam{{"1m", "5m", 300}, {"5m", "15m", 900}, {"15m", "30m", 1800}, {"30m", "1h", 3600}, {"1h", "4h", 14400}, {"4h", "1d", 86400}, {"1d", "1w", 604800}}
