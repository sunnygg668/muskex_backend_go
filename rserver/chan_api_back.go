package rserver

/*
grpc-web connect协议生成的js端调用，不支持关闭server端流，暂时停用流功能部分

	{
	  rpc Greet(NullMsg) returns (stream GreetResponse) {}
	  //StringParam 设置为 coinName: BTC ETH
	  rpc TradeChan(StringParam) returns (stream JsonBsResonse) {}
	  //StringParam 设置为 coinName: BTC ETH
	  rpc TickerChan(StringParam) returns (stream JsonBsResonse) {}
	  //StringParam 设置为 coinName: BTC ETH
	  rpc DepthChan(StringParam) returns (stream DepthData) {}
	}


	message GreetResponse {
	  string message = 1;
	}
*/
//
//type ChanServer struct {
//}
//
//func (as ChanServer) TickerChan(ctx context.Context, req *connect.Request[proto.StringParam], sconn *connect.ServerStream[proto.JsonBsResonse]) error {
//	coinName := strings.TrimSuffix(req.Msg.Str, "USDT")
//	depthSubMap := CoinTickerSubMap[coinName]
//	uid := shortuuid.New()
//	depthSubMap.Store(uid, &CommonSub{Uuid: uid, Ctx: ctx, Conn: sconn})
//
//	log.Print("TikerChan", uid)
//	<-ctx.Done()
//	log.Print("TikerChan ctx.Done", time.Now())
//	return errors.New("context.Done")
//}
//func (as ChanServer) TradeChan(ctx context.Context, req *connect.Request[proto.StringParam], sconn *connect.ServerStream[proto.JsonBsResonse]) error {
//	coinName := strings.TrimSuffix(req.Msg.Str, "USDT")
//	depthSubMap := CoinTradeSubMap[coinName]
//	uid := shortuuid.New()
//	depthSubMap.Store(uid, &CommonSub{Uuid: uid, Ctx: ctx, Conn: sconn})
//
//	log.Print("TradeChan", uid)
//	<-ctx.Done()
//	log.Print("TradeChan ctx.Done", time.Now())
//	return errors.New("context.Done")
//}
//func (as ChanServer) DepthChan(ctx context.Context, req *connect.Request[proto.StringParam], sconn *connect.ServerStream[proto.DepthChanReonse]) error {
//	coinName := strings.TrimSuffix(req.Msg.Str, "USDT")
//	depthSubMap := CoinDepthSubMap[coinName]
//	uid := shortuuid.New()
//	depthSubMap.Store(uid, &DepthSub{Uuid: uid, Ctx: ctx, Conn: sconn})
//
//	log.Print("DepthChan", uid)
//	<-ctx.Done()
//	log.Print("DepthChan ctx.Done", time.Now())
//	return errors.New("context.Done")
//}
//func (as ChanServer) Greet(ctx context.Context, req *connect.Request[proto.NullMsg], sconn *connect.ServerStream[proto.GreetResponse]) error {
//	after := time.After(511 * time.Second)
//	now := time.Now().String()
//	for {
//		// after 5 seconds, return the function
//		select {
//		case <-after:
//			log.Print("return")
//			return errors.New("asdb")
//		case <-ctx.Done():
//			log.Print("ctx.Done", now)
//			return errors.New("asdb")
//			return nil
//		default:
//			log.Print("begin time", now)
//			err := sconn.Send(&proto.GreetResponse{Message: "hello" + time.Now().String()})
//			if err != nil {
//				log.Print("Send err", err)
//				return err
//			}
//			time.Sleep(1 * time.Second)
//		}
//
//	}
//}
//
//func StartChans() {
//	initCoinSubMap()
//	go pubDepthChan()
//	go pubTradeChan()
//	go pubTickerChan()
//}
//func initCoinSubMap() {
//	coinItems := []*model.Coin{}
//	if err := utils.RawOrm.Find(&coinItems).Error; err != nil {
//		log.Fatal(err)
//	}
//	for _, item := range coinItems {
//		CoinDepthSubMap[item.Name] = &sync.Map{}
//		CoinTradeSubMap[item.Name] = &sync.Map{}
//		CoinTickerSubMap[item.Name] = &sync.Map{}
//	}
//}
//
//// map key is coinNmae; value of syncMap is (uuid,DepthSub)
//var CoinDepthSubMap = map[string]*sync.Map{}
//var CoinTradeSubMap = map[string]*sync.Map{}
//var CoinTickerSubMap = map[string]*sync.Map{}
//
//type DepthSub struct {
//	Uuid string
//	Ctx  context.Context
//	Conn *connect.ServerStream[proto.DepthChanReonse]
//}
//type CommonSub struct {
//	Uuid string
//	Ctx  context.Context
//	Conn *connect.ServerStream[proto.JsonBsResonse]
//}
//
//func pubDepthChan() {
//	sendCount := 0
//	msgCount := 0
//	for {
//		msgCount++
//		event := <-DepthChan
//		coinName := strings.TrimSuffix(event.Symbol, "USDT")
//		depthSubMap := CoinDepthSubMap[coinName]
//		pcount := 0
//		var sendData *proto.DepthChanReonse
//		depthSubMap.Range(func(key, value interface{}) bool {
//			if sendData == nil {
//				bids := make([]*proto.PriceLevel, len(event.Bids))
//				asks := make([]*proto.PriceLevel, len(event.Asks))
//				for i := 0; i < len(event.Bids); i++ {
//					bids[i] = &proto.PriceLevel{Price: event.Bids[i].Price, Quantity: event.Bids[i].Quantity}
//					asks[i] = &proto.PriceLevel{Price: event.Asks[i].Price, Quantity: event.Asks[i].Quantity}
//				}
//				sendData = &proto.DepthChanReonse{
//					Symbol: event.Symbol,
//					Bids:   bids,
//					Asks:   asks,
//				}
//			}
//			sub := value.(*DepthSub)
//			// when su.Ctx is blocked, send message; if not, remove it from depthSubMap
//			select {
//			case <-sub.Ctx.Done():
//				depthSubMap.Delete(key)
//				return true
//			default:
//				err := sub.Conn.Send(sendData)
//				sendCount++
//				pcount++
//				if err != nil {
//					depthSubMap.Delete(key)
//				}
//			}
//			return true
//		})
//		if sendCount > 0 && sendCount%100 == 0 {
//			log.Println("pubTradeChan", "sendCount:", sendCount, "msgCount", msgCount, coinName, "p:", pcount)
//		}
//	}
//}
//func pubTradeChan() {
//	sendCount := 0
//	msgCount := 0
//	for {
//		event := <-TradeChan
//		msgCount++
//		pcount := 0
//		coinName := strings.TrimSuffix(event.Symbol, "USDT")
//		tradeSubMap := CoinTradeSubMap[coinName]
//		tradeSubMap.Range(func(key, value interface{}) bool {
//			sub := value.(*CommonSub)
//			select {
//			case <-sub.Ctx.Done():
//				tradeSubMap.Delete(key)
//				return true
//			default:
//				bs, _ := json.Marshal(event)
//				err := sub.Conn.Send(&proto.JsonBsResonse{Data: bs})
//				sendCount++
//				pcount++
//				if err != nil {
//					tradeSubMap.Delete(key)
//				}
//			}
//			return true
//		})
//		if sendCount > 0 && sendCount%100 == 0 {
//			log.Println("pubTradeChan", "sendCount:", sendCount, "msgCount", msgCount, coinName, "p:", pcount)
//		}
//	}
//}
//func pubTickerChan() {
//	for {
//		event := <-TickerChan
//		coinName := strings.TrimSuffix(event.Symbol, "USDT")
//		tickerSubMap := CoinTickerSubMap[coinName]
//		sendCount := 0
//		tickerSubMap.Range(func(key, value interface{}) bool {
//			sub := value.(*CommonSub)
//			select {
//			case <-sub.Ctx.Done():
//				tickerSubMap.Delete(key)
//				return true
//			default:
//				bs, _ := json.Marshal(event)
//				err := sub.Conn.Send(&proto.JsonBsResonse{Data: bs})
//				sendCount++
//				if err != nil {
//					tickerSubMap.Delete(key)
//				}
//			}
//			return true
//		})
//		if sendCount > 0 && sendCount%100 == 0 {
//			log.Println("pubTickerChan", coinName, "sendCount:", sendCount)
//		}
//	}
//}
