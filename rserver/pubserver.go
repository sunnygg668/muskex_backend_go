package rserver

import (
	"connectrpc.com/connect"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	binance_connector "github.com/binance/binance-connector-go"
	"gorm.io/gorm"
	"log"
	"muskex/gen/mproto"
	"muskex/gen/mproto/model"
	"muskex/utils"
	"strconv"
	"strings"
	"time"
)

type PubServer struct{}

//	func DbContext(ctx context.Context) *gorm.DB {
//		return utils.Cdb(ctx)
//	}
func (PubServer) Index(ctx context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.IndexResonse], error) {
	//log.Print("userId", proto.GetCtxUserId(ctx))
	res, err := LoadHomeIndexFromCache()
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res), nil
}

func (PubServer) MinerList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.MinerListResonses], error) {
	items := []*model.Miners{}
	err := utils.Orm.Find(&items).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.MinerListResonses{List: items}), nil
}
func (PubServer) CoinManagementList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.CoinManagementListResonses], error) {
	items := []*model.CoinManagement{}
	err := utils.Orm.Find(&items).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.CoinManagementListResonses{List: items}), nil
}
func (PubServer) CarouselList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.CarouselListResonse], error) {
	items := []*model.MarketCarousel{}
	err := utils.Orm.Order("weigh desc").Find(&items, "status=?", 1).Error
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if len(item.Url) > 0 {
			item.CanOpen = 1
		}
	}
	return connect.NewResponse(&mproto.CarouselListResonse{List: items}), nil
}

func (PubServer) BankList(ctx context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.BankListResonse], error) {
	items := []*model.FinancialBank{}
	err := utils.Orm.Order("id desc").Find(&items, "status=?", 1).Error
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.BankListResonse{List: items}), nil
}

// 按卡号获取卡信息。 对应旧项目的  /api/financial_card/getBank?card=xxxxxx
func (PubServer) GetBankByPre(c context.Context, req *connect.Request[mproto.GetBankByPreRequest]) (*connect.Response[model.FinancialBank], error) {
	length := len(req.Msg.Card)

	if length >= 10 {
		bank, err := GetBankByCode(req.Msg.Card, 10)
		if err == nil {
			return connect.NewResponse(bank), nil
		}
	}

	if length >= 9 {
		bank, err := GetBankByCode(req.Msg.Card, 9)
		if err == nil {
			return connect.NewResponse(bank), nil
		}
	}
	if length >= 8 {
		bank, err := GetBankByCode(req.Msg.Card, 8)
		if err == nil {
			return connect.NewResponse(bank), nil
		}
	}

	if length < 6 {
		return nil, fmt.Errorf("Card No error,length need large than 5")
	}

	bank, err := GetBankByCode(req.Msg.Card, 6)
	if err == nil {
		return connect.NewResponse(bank), nil
	} else {
		return nil, err
	}
}

// 源api /api/index/helpCenter
func (as PubServer) HelpList(c context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.HelpListResponse], error) {
	items := LoadDbConfig().Groups["help_center"]
	return connect.NewResponse(&mproto.HelpListResponse{List: items}), nil
}

// 源api /api/index/helpDetail?name=xxx
func (as PubServer) HelpDetail(c context.Context, req *connect.Request[mproto.StringParam]) (*connect.Response[model.Config], error) {
	item := LoadDbConfig().Names[req.Msg.Str]
	return connect.NewResponse(item), nil
}

// coin行情 对应旧项目的  /api/coin_data/tickerInfo
func (as PubServer) TickerInfo(ctx context.Context, req *connect.Request[mproto.StringParam]) (*connect.Response[mproto.TickerInfoResponse], error) {
	key := strings.TrimSuffix(req.Msg.Str, "/USDT")
	//bs := []byte{}
	var ticker *binance_connector.WsMarketTickerStatEvent
	v, ok := TickerMap.Load(key)
	if ok {
		ticker = v.(*binance_connector.WsMarketTickerStatEvent)
	}
	bs, _ := json.Marshal(ticker)
	return connect.NewResponse(&mproto.TickerInfoResponse{Data: bs}), nil
}

func (as PubServer) KlineTradeList(ctx context.Context, req *connect.Request[mproto.StringParam]) (*connect.Response[mproto.KlineTradeListResonse], error) {
	coinName := strings.TrimSuffix(req.Msg.Str, "/USDT")
	res, err := utils.CacheFromLru(1, "KlineTrade_"+coinName, 1, func() (interface{}, error) {
		es := []*binance_connector.WsAggTradeEvent{}
		v, ok := TradeMap.Load(coinName)
		if ok {
			es = v.([]*binance_connector.WsAggTradeEvent)
		}
		list := []*mproto.TradeEvent{}
		for _, e := range es {
			list = append(list, &mproto.TradeEvent{
				Price:    e.Price,
				Quantity: e.Quantity,
				//Time:         e.Time / 1000,
				TradeTime:    e.TradeTime / 1000,
				IsBuyerMaker: e.IsBuyerMaker,
			})
		}
		return list, nil
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.KlineTradeListResonse{List: res.([]*mproto.TradeEvent)}), nil
}

func (as PubServer) KlineDepth(ctx context.Context, req *connect.Request[mproto.StringParam]) (*connect.Response[mproto.JsonBsResonse], error) {
	coinName := strings.TrimSuffix(req.Msg.Str, "/USDT")
	event := &binance_connector.WsPartialDepthEvent{}
	v, ok := DepthMap.Load(coinName)
	if ok {
		event = v.(*binance_connector.WsPartialDepthEvent)
	}
	bs1, _ := json.Marshal(event)
	return connect.NewResponse(&mproto.JsonBsResonse{Data: bs1}), nil
}

// TickerInfo coin行情 对应旧项目的,同时附了depth ticker数据  /api/coin_data/kline
func (as PubServer) KlineInfoLast(ctx context.Context, req *connect.Request[mproto.KlineInfoRequest]) (*connect.Response[mproto.KlineInfoLastResonse], error) {
	coinName := strings.TrimSuffix(req.Msg.KlineType, "/USDT")
	res, err := utils.CacheFromLru(1, "KlineInfoNews_"+coinName+req.Msg.Interval, 2, func() (interface{}, error) {
		kline := model.KlineData{}
		utils.Orm.Order("start_time desc").First(&kline, "symbol=? and stat_type=?", coinName, req.Msg.Interval)
		var depth *mproto.DepthData
		v, ok := DepthMap.Load(coinName)
		if ok {
			depthEvnent := v.(*binance_connector.WsPartialDepthEvent)
			bids := make([]*mproto.PriceLevel, len(depthEvnent.Bids))
			asks := make([]*mproto.PriceLevel, len(depthEvnent.Asks))
			for i := 0; i < len(depthEvnent.Bids); i++ {
				bids[i] = &mproto.PriceLevel{Price: depthEvnent.Bids[i].Price, Quantity: depthEvnent.Bids[i].Quantity}
				asks[i] = &mproto.PriceLevel{Price: depthEvnent.Asks[i].Price, Quantity: depthEvnent.Asks[i].Quantity}
			}
			depth = &mproto.DepthData{
				Bids: bids,
				Asks: asks,
			}
		}
		var tickerEvent *binance_connector.WsMarketTickerStatEvent
		var ticker *model.TickerData
		v1, ok1 := TickerMap.Load(coinName)
		if ok1 {
			tickerEvent = v1.(*binance_connector.WsMarketTickerStatEvent)
			open, _ := strconv.ParseFloat(tickerEvent.OpenPrice, 64)
			price, _ := strconv.ParseFloat(tickerEvent.LastPrice, 64)
			high, _ := strconv.ParseFloat(tickerEvent.HighPrice, 64)
			low, _ := strconv.ParseFloat(tickerEvent.LowPrice, 64)
			change, _ := strconv.ParseFloat(tickerEvent.PriceChangePercent, 64)
			total, _ := strconv.ParseFloat(tickerEvent.QuoteVolume, 64)
			ticker = &model.TickerData{
				Open:               open,
				Price:              price,
				High:               high,
				Low:                low,
				QuoteVolume:        total,
				PriceChangePercent: change,
			}
		}
		return &mproto.KlineInfoLastResonse{
			Kline:  &kline,
			Depth:  depth,
			Ticker: ticker,
		}, nil
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(res.(*mproto.KlineInfoLastResonse)), nil
}
func (as PubServer) KlineInfoList(ctx context.Context, req *connect.Request[mproto.KlineInfoRequest]) (*connect.Response[mproto.KlineDataResponse], error) {
	coinName := strings.TrimSuffix(req.Msg.KlineType, "/USDT")
	res, err := utils.CacheFromLru(1, "KlineInfoList_"+coinName+"_"+req.Msg.Interval, 2, func() (interface{}, error) {
		list := []*model.KlineData{}
		utils.Orm.Find(&list, "symbol=? and stat_type=?", coinName, req.Msg.Interval)
		return list, nil
	})
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&mproto.KlineDataResponse{List: res.([]*model.KlineData)}), nil
}

func (as PubServer) IndexAllCoin(ctx context.Context, req *connect.Request[mproto.PidParam]) (*connect.Response[mproto.IndexAllCoinResponse], error) {
	if req.Msg.Pid == 0 {
		req.Msg.Pid = 1
	}
	pid := req.Msg.Pid
	res, _ := utils.CacheFromLru(1, "IndexAllCoin"+strconv.Itoa(int(pid)), 2, func() (interface{}, error) {
		ranks := &mproto.Ranks{}
		err := utils.Orm.Raw(sql_coinid, pid).Scan(ranks).Error
		if err != nil {
			return nil, err
		}
		coins := &[]*mproto.RankItem{}
		utils.Orm.Raw(sql_coins, pid, pid).Find(coins)
		return &mproto.IndexAllCoinResponse{Coins: *coins, Ranks: ranks}, nil
	})
	return connect.NewResponse(res.(*mproto.IndexAllCoinResponse)), nil
}

var sql_coinid = `
with base_coin as
         (select t.id, price_change_percent, total_trade
          from ba_rank_item t
                   inner join ba_p_coin t1 on t.name = t1.name
          where t1.pid = ?),
     rank_bull as
         (select id from base_coin order by price_change_percent desc limit 8),
     rank_bear as
         (select id from base_coin order by price_change_percent asc limit 8),
     rank_trade as
             (select id from base_coin order by total_trade desc limit 8)
select (select group_concat(id) ids
        from rank_bull)  bull,
       (select group_concat(id) ids
        from rank_bear)  bear,
       (select group_concat(id) ids
        from rank_trade) trade,
       (select group_concat(id) from ba_rank_item where name in('BTC','ETH','BNB')) as  recommand`

var sql_coins = `
with base_coin as
         (select t.id, price_change_percent, total_trade
          from ba_rank_item t
                   inner join ba_p_coin t1 on t.name = t1.name
          where t1.pid = ?),
     rank_bull as
         (select id from base_coin order by price_change_percent desc limit 8),
     rank_bear as
         (select id from base_coin order by price_change_percent asc limit 8),
     rank_trade as
             (select * from base_coin order by total_trade desc limit 8)
select t.id,
       t.name,
       t.alias,
       '' logo_image,
       t.initial_price,
       t.price_change_percent,
       t.total_trade,
       t1.margin
from ba_rank_item t
         inner join(select id
                    from rank_bull
                    union
                    select id
                    from rank_bear
                    union
                    select id
                    from rank_trade
                    union
                    select id from ba_rank_item where name in('BTC','ETH','BNB')) aa on t.id = aa.id
         inner join ba_p_coin t1 on t.name = t1.name and t1.pid = ?
;
`

// 排行榜
func (as PubServer) RankList(ctx context.Context, req *connect.Request[mproto.RankListRequest]) (*connect.Response[mproto.RankListResponse], error) {
	res, _ := utils.CacheFromLru(1, "rank_list"+req.Msg.RankType.String()+"pid"+strconv.Itoa(int(req.Msg.Pid)), 2, func() (interface{}, error) {
		orderstr := ""
		if req.Msg.RankType == mproto.RankListRequest_RANK_TYPE_DESC {
			orderstr = "price_change_percent desc"
		} else if req.Msg.RankType == mproto.RankListRequest_RANK_TYPE_ASC {
			orderstr = "price_change_percent asc"
		} else if req.Msg.RankType == mproto.RankListRequest_RANK_TYPE_TRADE {
			orderstr = "total_trade desc"
		} else if req.Msg.RankType == mproto.RankListRequest_RANK_TYPE_MARGIN_ASC {
			orderstr = "margin"
		} else if req.Msg.RankType == mproto.RankListRequest_RANK_TYPE_MARGIN_DESC {
			orderstr = "margin desc"
		}
		items := []*mproto.RankItem{}
		sql := `
select t.id,
       t.name,
       t.alias,
       t.initial_price,
       t.total_trade,
       t.price_change_percent,
       t1.margin
from ba_rank_item t
         inner join ba_p_coin t1 on t.name = t1.name
where t1.pid = ? order by ` + orderstr
		if req.Msg.Pid == 0 {
			req.Msg.Pid = 1
		}
		utils.Orm.Raw(sql, req.Msg.Pid).Scan(&items)

		sql = `with coins as
   ( select  id,name
      from ba_p_coin
      where is_hot = 1 and pid = ?
      order by margin desc limit 10)
select group_concat(name)
from coins;`
		hotIds := ""
		utils.Orm.Raw(sql, req.Msg.Pid).Scan(&hotIds)
		return &mproto.RankListResponse{List: items, HotNames: hotIds}, nil
	})
	return connect.NewResponse(res.(*mproto.RankListResponse)), nil
}

// 发送短信
func (as PubServer) SendSms(ctx context.Context, req *connect.Request[mproto.SendSmsRequest]) (*connect.Response[mproto.MsgResponse], error) {
	if req.Msg.TemplateCode == "" {
		return nil, fmt.Errorf("参数错误")
	}
	if IsValidPhoneNumber(req.Msg.Mobile) == false {
		return nil, fmt.Errorf("手机号码格式错误")
	}

	tuser := &model.User{}
	err := utils.Orm.First(tuser, "mobile=?", req.Msg.Mobile).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		if In_array(req.Msg.TemplateCode, []string{"user_retrieve_pwd", "user_retrieve_fund_pwd", "user_mobile_verify", "user_change_mobile_old", "user_login", "user_change_pwd"}) {
			return nil, errors.New("Mobile number not registered")
		}
	} else if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if req.Msg.TemplateCode == "user_register" {
		return nil, errors.New("Mobile number has been registered, please log in directly")
	} else if req.Msg.TemplateCode == "user_change_mobile" {
		return nil, errors.New("The mobile number has been occupied")
	}

	templateData := &model.SmsTemplate{}
	terr := utils.Orm.First(templateData, "code = ? and status=1", req.Msg.TemplateCode).Error
	if terr != nil && terr == gorm.ErrRecordNotFound {
		return nil, errors.New("SMS template does not exist")
	} else if terr != nil {
		return nil, terr
	} else if templateData.Template == "" && templateData.Content == "" {
		return nil, errors.New("SMS template error")
	}

	cfg := LoadDbConfig().NameValues
	InitializeSendSmsReqConfig(cfg["sms_account"], cfg["sms_pswd"], "http://mxthk.weiwebs.cn/msg/HttpBatchSendSM") // 国内地址https://www.weiwebs.cn/msg/HttpBatchSendSM
	template := AnalysisVariable(templateData.Content, templateData.Variables, req.Msg.Mobile+req.Msg.TemplateCode, map[string]string{})
	TemplateAnalysisAfter(template, map[string]string{"mobile": req.Msg.Mobile, "template_code": req.Msg.TemplateCode})
	sendSmsResponse, _ := SendSms(req.Msg.Mobile, &sendSmsMessage{template: templateData.Template, content: template["content"].(string), data: template["variables"].(map[string]string)})

	return connect.NewResponse(&mproto.MsgResponse{Message: "短信发送成功" + sendSmsResponse.Message}), nil
}

func (as PubServer) Greet(ctx context.Context, req *connect.Request[mproto.StringParam], sconn *connect.ServerStream[mproto.GreetResponse]) error {
	after := time.After(511 * time.Second)
	now := time.Now().String()
	for {
		// after 5 seconds, return the function
		select {
		case <-after:
			log.Print("return")
			return errors.New("asdb")
		case <-ctx.Done():
			log.Print("ctx.Done", now)
			return errors.New("asdb")
			return nil
		default:
			log.Print("begin time", now)
			err := sconn.Send(&mproto.GreetResponse{Message: "hello" + req.Msg.Str + time.Now().String()})
			if err != nil {
				log.Print("Send err", err)
				return err
			}
			time.Sleep(1 * time.Second)
		}

	}
}

//func (protoServer) GetMemberInfo(context.Context, *connect.Request[proto.GetMemberInfoReqest]) (*connect.Response[proto.IndexResonse], error) {
//
//}
//func walletLogin(req *connect.Request[proto.GetMemberInfoReqest]) (*proto.GetMemberInfoResponse, error) {
//	walletAddr := req.Msg.WalletAddr
//	user := new(model.User)
//	token := new(model.Token)
//
//	err := utils.Orm.First(user, model.User{WalletAddr: walletAddr}).Error
//	if err != nil {
//		return nil, err
//	}
//	if user.Status == "0" {
//		return nil, errors.New("Account disabled")
//	}
//	if user.LoginFailure >= 10 && time.Now().Unix()-user.LastLoginTime < 86400 {
//		return nil, errors.New("Please try again after 1 day")
//	}
//	//if user.Password != pwdSum(lreq.Password, user.Salt) {
//	//	controls.ResErrMsg(c, "Password is incorrect")
//	//	return
//	//}
//	user.LoginFailure = 0
//	user.LastLoginTime = time.Now().Unix()
//	user.LastLoginIp = req.Peer().Addr
//	utils.Orm.Save(user)
//
//}
//func pwdSum(pwd, salt string) string {
//	return utils.GetMd5String(utils.GetMd5String(pwd) + salt)
//}
