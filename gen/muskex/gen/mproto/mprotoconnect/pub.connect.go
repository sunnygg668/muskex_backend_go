// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: mproto/pub.proto

package mprotoconnect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	mproto "muskex/gen/mproto"
	model "muskex/gen/mproto/model"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// PubServiceName is the fully-qualified name of the PubService service.
	PubServiceName = "mproto.PubService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// PubServiceIndexProcedure is the fully-qualified name of the PubService's Index RPC.
	PubServiceIndexProcedure = "/mproto.PubService/Index"
	// PubServiceHelpListProcedure is the fully-qualified name of the PubService's HelpList RPC.
	PubServiceHelpListProcedure = "/mproto.PubService/HelpList"
	// PubServiceHelpDetailProcedure is the fully-qualified name of the PubService's HelpDetail RPC.
	PubServiceHelpDetailProcedure = "/mproto.PubService/HelpDetail"
	// PubServiceCarouselListProcedure is the fully-qualified name of the PubService's CarouselList RPC.
	PubServiceCarouselListProcedure = "/mproto.PubService/CarouselList"
	// PubServiceIndexAllCoinProcedure is the fully-qualified name of the PubService's IndexAllCoin RPC.
	PubServiceIndexAllCoinProcedure = "/mproto.PubService/IndexAllCoin"
	// PubServiceBankListProcedure is the fully-qualified name of the PubService's BankList RPC.
	PubServiceBankListProcedure = "/mproto.PubService/BankList"
	// PubServiceCoinManagementListProcedure is the fully-qualified name of the PubService's
	// CoinManagementList RPC.
	PubServiceCoinManagementListProcedure = "/mproto.PubService/CoinManagementList"
	// PubServiceMinerListProcedure is the fully-qualified name of the PubService's MinerList RPC.
	PubServiceMinerListProcedure = "/mproto.PubService/MinerList"
	// PubServiceGetBankByPreProcedure is the fully-qualified name of the PubService's GetBankByPre RPC.
	PubServiceGetBankByPreProcedure = "/mproto.PubService/GetBankByPre"
	// PubServiceKlineInfoListProcedure is the fully-qualified name of the PubService's KlineInfoList
	// RPC.
	PubServiceKlineInfoListProcedure = "/mproto.PubService/KlineInfoList"
	// PubServiceKlineInfoLastProcedure is the fully-qualified name of the PubService's KlineInfoLast
	// RPC.
	PubServiceKlineInfoLastProcedure = "/mproto.PubService/KlineInfoLast"
	// PubServiceKlineTradeListProcedure is the fully-qualified name of the PubService's KlineTradeList
	// RPC.
	PubServiceKlineTradeListProcedure = "/mproto.PubService/KlineTradeList"
	// PubServiceRankListProcedure is the fully-qualified name of the PubService's RankList RPC.
	PubServiceRankListProcedure = "/mproto.PubService/RankList"
	// PubServiceSendSmsProcedure is the fully-qualified name of the PubService's SendSms RPC.
	PubServiceSendSmsProcedure = "/mproto.PubService/SendSms"
	// PubServiceGreetProcedure is the fully-qualified name of the PubService's Greet RPC.
	PubServiceGreetProcedure = "/mproto.PubService/Greet"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	pubServiceServiceDescriptor                  = mproto.File_mproto_pub_proto.Services().ByName("PubService")
	pubServiceIndexMethodDescriptor              = pubServiceServiceDescriptor.Methods().ByName("Index")
	pubServiceHelpListMethodDescriptor           = pubServiceServiceDescriptor.Methods().ByName("HelpList")
	pubServiceHelpDetailMethodDescriptor         = pubServiceServiceDescriptor.Methods().ByName("HelpDetail")
	pubServiceCarouselListMethodDescriptor       = pubServiceServiceDescriptor.Methods().ByName("CarouselList")
	pubServiceIndexAllCoinMethodDescriptor       = pubServiceServiceDescriptor.Methods().ByName("IndexAllCoin")
	pubServiceBankListMethodDescriptor           = pubServiceServiceDescriptor.Methods().ByName("BankList")
	pubServiceCoinManagementListMethodDescriptor = pubServiceServiceDescriptor.Methods().ByName("CoinManagementList")
	pubServiceMinerListMethodDescriptor          = pubServiceServiceDescriptor.Methods().ByName("MinerList")
	pubServiceGetBankByPreMethodDescriptor       = pubServiceServiceDescriptor.Methods().ByName("GetBankByPre")
	pubServiceKlineInfoListMethodDescriptor      = pubServiceServiceDescriptor.Methods().ByName("KlineInfoList")
	pubServiceKlineInfoLastMethodDescriptor      = pubServiceServiceDescriptor.Methods().ByName("KlineInfoLast")
	pubServiceKlineTradeListMethodDescriptor     = pubServiceServiceDescriptor.Methods().ByName("KlineTradeList")
	pubServiceRankListMethodDescriptor           = pubServiceServiceDescriptor.Methods().ByName("RankList")
	pubServiceSendSmsMethodDescriptor            = pubServiceServiceDescriptor.Methods().ByName("SendSms")
	pubServiceGreetMethodDescriptor              = pubServiceServiceDescriptor.Methods().ByName("Greet")
)

// PubServiceClient is a client for the mproto.PubService service.
type PubServiceClient interface {
	// 首页，对应旧项目的 /api/index/home
	Index(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.IndexResonse], error)
	// 源api /api/index/helpCenter
	HelpList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.HelpListResponse], error)
	// 源api /api/index/helpDetail?name=xxx
	HelpDetail(context.Context, *connect.Request[mproto.StringParam]) (*connect.Response[model.Config], error)
	// 首页轮播 对应旧项目的 /api/index/carouselList
	CarouselList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.CarouselListResonse], error)
	// 首页3个币排行及推荐，币数据中的logImage路径，使用这样的模板 https://image.tecajx.vipimages/{xxx}.png ; 如BTC使用https://image.tecajx.vip/images/BTC.png
	IndexAllCoin(context.Context, *connect.Request[mproto.PidParam]) (*connect.Response[mproto.IndexAllCoinResponse], error)
	// 银行卡列表 对应旧项目的 /api/financial_card/bankList
	BankList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.BankListResonse], error)
	// usdt理财 对应旧项目的 /coin_management/index
	CoinManagementList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.CoinManagementListResonses], error)
	// 矿机列表
	MinerList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.MinerListResonses], error)
	// 按卡号获取卡信息。 对应旧项目的  /api/financial_card/getBank?card=xxxxxx
	GetBankByPre(context.Context, *connect.Request[mproto.GetBankByPreRequest]) (*connect.Response[model.FinancialBank], error)
	// kline初始列表:服务端缓存5秒 对应旧项目的  /api/coin_data/kline
	KlineInfoList(context.Context, *connect.Request[mproto.KlineInfoRequest]) (*connect.Response[mproto.KlineDataResponse], error)
	// kline离当前时间最近的1条信息:服务端缓存1秒; 同时附带了depth,ticker数据。 对应旧项目的  /api/coin_data/kline
	KlineInfoLast(context.Context, *connect.Request[mproto.KlineInfoRequest]) (*connect.Response[mproto.KlineInfoLastResonse], error)
	// kline trade信息:服务端缓存1秒。 对应旧项目的  /api/coin_data/trade
	KlineTradeList(context.Context, *connect.Request[mproto.StringParam]) (*connect.Response[mproto.KlineTradeListResonse], error)
	// 行情列表  对应旧项目的  api/coin_data/ticker
	RankList(context.Context, *connect.Request[mproto.RankListRequest]) (*connect.Response[mproto.RankListResponse], error)
	SendSms(context.Context, *connect.Request[mproto.SendSmsRequest]) (*connect.Response[mproto.MsgResponse], error)
	Greet(context.Context, *connect.Request[mproto.StringParam]) (*connect.ServerStreamForClient[mproto.GreetResponse], error)
}

// NewPubServiceClient constructs a client for the mproto.PubService service. By default, it uses
// the Connect protocol with the binary Protobuf Codec, asks for gzipped responses, and sends
// uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the connect.WithGRPC() or
// connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewPubServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) PubServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &pubServiceClient{
		index: connect.NewClient[mproto.NullMsg, mproto.IndexResonse](
			httpClient,
			baseURL+PubServiceIndexProcedure,
			connect.WithSchema(pubServiceIndexMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		helpList: connect.NewClient[mproto.NullMsg, mproto.HelpListResponse](
			httpClient,
			baseURL+PubServiceHelpListProcedure,
			connect.WithSchema(pubServiceHelpListMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		helpDetail: connect.NewClient[mproto.StringParam, model.Config](
			httpClient,
			baseURL+PubServiceHelpDetailProcedure,
			connect.WithSchema(pubServiceHelpDetailMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		carouselList: connect.NewClient[mproto.NullMsg, mproto.CarouselListResonse](
			httpClient,
			baseURL+PubServiceCarouselListProcedure,
			connect.WithSchema(pubServiceCarouselListMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		indexAllCoin: connect.NewClient[mproto.PidParam, mproto.IndexAllCoinResponse](
			httpClient,
			baseURL+PubServiceIndexAllCoinProcedure,
			connect.WithSchema(pubServiceIndexAllCoinMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		bankList: connect.NewClient[mproto.NullMsg, mproto.BankListResonse](
			httpClient,
			baseURL+PubServiceBankListProcedure,
			connect.WithSchema(pubServiceBankListMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		coinManagementList: connect.NewClient[mproto.NullMsg, mproto.CoinManagementListResonses](
			httpClient,
			baseURL+PubServiceCoinManagementListProcedure,
			connect.WithSchema(pubServiceCoinManagementListMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		minerList: connect.NewClient[mproto.NullMsg, mproto.MinerListResonses](
			httpClient,
			baseURL+PubServiceMinerListProcedure,
			connect.WithSchema(pubServiceMinerListMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		getBankByPre: connect.NewClient[mproto.GetBankByPreRequest, model.FinancialBank](
			httpClient,
			baseURL+PubServiceGetBankByPreProcedure,
			connect.WithSchema(pubServiceGetBankByPreMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		klineInfoList: connect.NewClient[mproto.KlineInfoRequest, mproto.KlineDataResponse](
			httpClient,
			baseURL+PubServiceKlineInfoListProcedure,
			connect.WithSchema(pubServiceKlineInfoListMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		klineInfoLast: connect.NewClient[mproto.KlineInfoRequest, mproto.KlineInfoLastResonse](
			httpClient,
			baseURL+PubServiceKlineInfoLastProcedure,
			connect.WithSchema(pubServiceKlineInfoLastMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		klineTradeList: connect.NewClient[mproto.StringParam, mproto.KlineTradeListResonse](
			httpClient,
			baseURL+PubServiceKlineTradeListProcedure,
			connect.WithSchema(pubServiceKlineTradeListMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		rankList: connect.NewClient[mproto.RankListRequest, mproto.RankListResponse](
			httpClient,
			baseURL+PubServiceRankListProcedure,
			connect.WithSchema(pubServiceRankListMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		sendSms: connect.NewClient[mproto.SendSmsRequest, mproto.MsgResponse](
			httpClient,
			baseURL+PubServiceSendSmsProcedure,
			connect.WithSchema(pubServiceSendSmsMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		greet: connect.NewClient[mproto.StringParam, mproto.GreetResponse](
			httpClient,
			baseURL+PubServiceGreetProcedure,
			connect.WithSchema(pubServiceGreetMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// pubServiceClient implements PubServiceClient.
type pubServiceClient struct {
	index              *connect.Client[mproto.NullMsg, mproto.IndexResonse]
	helpList           *connect.Client[mproto.NullMsg, mproto.HelpListResponse]
	helpDetail         *connect.Client[mproto.StringParam, model.Config]
	carouselList       *connect.Client[mproto.NullMsg, mproto.CarouselListResonse]
	indexAllCoin       *connect.Client[mproto.PidParam, mproto.IndexAllCoinResponse]
	bankList           *connect.Client[mproto.NullMsg, mproto.BankListResonse]
	coinManagementList *connect.Client[mproto.NullMsg, mproto.CoinManagementListResonses]
	minerList          *connect.Client[mproto.NullMsg, mproto.MinerListResonses]
	getBankByPre       *connect.Client[mproto.GetBankByPreRequest, model.FinancialBank]
	klineInfoList      *connect.Client[mproto.KlineInfoRequest, mproto.KlineDataResponse]
	klineInfoLast      *connect.Client[mproto.KlineInfoRequest, mproto.KlineInfoLastResonse]
	klineTradeList     *connect.Client[mproto.StringParam, mproto.KlineTradeListResonse]
	rankList           *connect.Client[mproto.RankListRequest, mproto.RankListResponse]
	sendSms            *connect.Client[mproto.SendSmsRequest, mproto.MsgResponse]
	greet              *connect.Client[mproto.StringParam, mproto.GreetResponse]
}

// Index calls mproto.PubService.Index.
func (c *pubServiceClient) Index(ctx context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.IndexResonse], error) {
	return c.index.CallUnary(ctx, req)
}

// HelpList calls mproto.PubService.HelpList.
func (c *pubServiceClient) HelpList(ctx context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.HelpListResponse], error) {
	return c.helpList.CallUnary(ctx, req)
}

// HelpDetail calls mproto.PubService.HelpDetail.
func (c *pubServiceClient) HelpDetail(ctx context.Context, req *connect.Request[mproto.StringParam]) (*connect.Response[model.Config], error) {
	return c.helpDetail.CallUnary(ctx, req)
}

// CarouselList calls mproto.PubService.CarouselList.
func (c *pubServiceClient) CarouselList(ctx context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.CarouselListResonse], error) {
	return c.carouselList.CallUnary(ctx, req)
}

// IndexAllCoin calls mproto.PubService.IndexAllCoin.
func (c *pubServiceClient) IndexAllCoin(ctx context.Context, req *connect.Request[mproto.PidParam]) (*connect.Response[mproto.IndexAllCoinResponse], error) {
	return c.indexAllCoin.CallUnary(ctx, req)
}

// BankList calls mproto.PubService.BankList.
func (c *pubServiceClient) BankList(ctx context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.BankListResonse], error) {
	return c.bankList.CallUnary(ctx, req)
}

// CoinManagementList calls mproto.PubService.CoinManagementList.
func (c *pubServiceClient) CoinManagementList(ctx context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.CoinManagementListResonses], error) {
	return c.coinManagementList.CallUnary(ctx, req)
}

// MinerList calls mproto.PubService.MinerList.
func (c *pubServiceClient) MinerList(ctx context.Context, req *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.MinerListResonses], error) {
	return c.minerList.CallUnary(ctx, req)
}

// GetBankByPre calls mproto.PubService.GetBankByPre.
func (c *pubServiceClient) GetBankByPre(ctx context.Context, req *connect.Request[mproto.GetBankByPreRequest]) (*connect.Response[model.FinancialBank], error) {
	return c.getBankByPre.CallUnary(ctx, req)
}

// KlineInfoList calls mproto.PubService.KlineInfoList.
func (c *pubServiceClient) KlineInfoList(ctx context.Context, req *connect.Request[mproto.KlineInfoRequest]) (*connect.Response[mproto.KlineDataResponse], error) {
	return c.klineInfoList.CallUnary(ctx, req)
}

// KlineInfoLast calls mproto.PubService.KlineInfoLast.
func (c *pubServiceClient) KlineInfoLast(ctx context.Context, req *connect.Request[mproto.KlineInfoRequest]) (*connect.Response[mproto.KlineInfoLastResonse], error) {
	return c.klineInfoLast.CallUnary(ctx, req)
}

// KlineTradeList calls mproto.PubService.KlineTradeList.
func (c *pubServiceClient) KlineTradeList(ctx context.Context, req *connect.Request[mproto.StringParam]) (*connect.Response[mproto.KlineTradeListResonse], error) {
	return c.klineTradeList.CallUnary(ctx, req)
}

// RankList calls mproto.PubService.RankList.
func (c *pubServiceClient) RankList(ctx context.Context, req *connect.Request[mproto.RankListRequest]) (*connect.Response[mproto.RankListResponse], error) {
	return c.rankList.CallUnary(ctx, req)
}

// SendSms calls mproto.PubService.SendSms.
func (c *pubServiceClient) SendSms(ctx context.Context, req *connect.Request[mproto.SendSmsRequest]) (*connect.Response[mproto.MsgResponse], error) {
	return c.sendSms.CallUnary(ctx, req)
}

// Greet calls mproto.PubService.Greet.
func (c *pubServiceClient) Greet(ctx context.Context, req *connect.Request[mproto.StringParam]) (*connect.ServerStreamForClient[mproto.GreetResponse], error) {
	return c.greet.CallServerStream(ctx, req)
}

// PubServiceHandler is an implementation of the mproto.PubService service.
type PubServiceHandler interface {
	// 首页，对应旧项目的 /api/index/home
	Index(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.IndexResonse], error)
	// 源api /api/index/helpCenter
	HelpList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.HelpListResponse], error)
	// 源api /api/index/helpDetail?name=xxx
	HelpDetail(context.Context, *connect.Request[mproto.StringParam]) (*connect.Response[model.Config], error)
	// 首页轮播 对应旧项目的 /api/index/carouselList
	CarouselList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.CarouselListResonse], error)
	// 首页3个币排行及推荐，币数据中的logImage路径，使用这样的模板 https://image.tecajx.vipimages/{xxx}.png ; 如BTC使用https://image.tecajx.vip/images/BTC.png
	IndexAllCoin(context.Context, *connect.Request[mproto.PidParam]) (*connect.Response[mproto.IndexAllCoinResponse], error)
	// 银行卡列表 对应旧项目的 /api/financial_card/bankList
	BankList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.BankListResonse], error)
	// usdt理财 对应旧项目的 /coin_management/index
	CoinManagementList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.CoinManagementListResonses], error)
	// 矿机列表
	MinerList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.MinerListResonses], error)
	// 按卡号获取卡信息。 对应旧项目的  /api/financial_card/getBank?card=xxxxxx
	GetBankByPre(context.Context, *connect.Request[mproto.GetBankByPreRequest]) (*connect.Response[model.FinancialBank], error)
	// kline初始列表:服务端缓存5秒 对应旧项目的  /api/coin_data/kline
	KlineInfoList(context.Context, *connect.Request[mproto.KlineInfoRequest]) (*connect.Response[mproto.KlineDataResponse], error)
	// kline离当前时间最近的1条信息:服务端缓存1秒; 同时附带了depth,ticker数据。 对应旧项目的  /api/coin_data/kline
	KlineInfoLast(context.Context, *connect.Request[mproto.KlineInfoRequest]) (*connect.Response[mproto.KlineInfoLastResonse], error)
	// kline trade信息:服务端缓存1秒。 对应旧项目的  /api/coin_data/trade
	KlineTradeList(context.Context, *connect.Request[mproto.StringParam]) (*connect.Response[mproto.KlineTradeListResonse], error)
	// 行情列表  对应旧项目的  api/coin_data/ticker
	RankList(context.Context, *connect.Request[mproto.RankListRequest]) (*connect.Response[mproto.RankListResponse], error)
	SendSms(context.Context, *connect.Request[mproto.SendSmsRequest]) (*connect.Response[mproto.MsgResponse], error)
	Greet(context.Context, *connect.Request[mproto.StringParam], *connect.ServerStream[mproto.GreetResponse]) error
}

// NewPubServiceHandler builds an HTTP handler from the service implementation. It returns the path
// on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewPubServiceHandler(svc PubServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	pubServiceIndexHandler := connect.NewUnaryHandler(
		PubServiceIndexProcedure,
		svc.Index,
		connect.WithSchema(pubServiceIndexMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceHelpListHandler := connect.NewUnaryHandler(
		PubServiceHelpListProcedure,
		svc.HelpList,
		connect.WithSchema(pubServiceHelpListMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceHelpDetailHandler := connect.NewUnaryHandler(
		PubServiceHelpDetailProcedure,
		svc.HelpDetail,
		connect.WithSchema(pubServiceHelpDetailMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceCarouselListHandler := connect.NewUnaryHandler(
		PubServiceCarouselListProcedure,
		svc.CarouselList,
		connect.WithSchema(pubServiceCarouselListMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceIndexAllCoinHandler := connect.NewUnaryHandler(
		PubServiceIndexAllCoinProcedure,
		svc.IndexAllCoin,
		connect.WithSchema(pubServiceIndexAllCoinMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceBankListHandler := connect.NewUnaryHandler(
		PubServiceBankListProcedure,
		svc.BankList,
		connect.WithSchema(pubServiceBankListMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceCoinManagementListHandler := connect.NewUnaryHandler(
		PubServiceCoinManagementListProcedure,
		svc.CoinManagementList,
		connect.WithSchema(pubServiceCoinManagementListMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceMinerListHandler := connect.NewUnaryHandler(
		PubServiceMinerListProcedure,
		svc.MinerList,
		connect.WithSchema(pubServiceMinerListMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceGetBankByPreHandler := connect.NewUnaryHandler(
		PubServiceGetBankByPreProcedure,
		svc.GetBankByPre,
		connect.WithSchema(pubServiceGetBankByPreMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceKlineInfoListHandler := connect.NewUnaryHandler(
		PubServiceKlineInfoListProcedure,
		svc.KlineInfoList,
		connect.WithSchema(pubServiceKlineInfoListMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceKlineInfoLastHandler := connect.NewUnaryHandler(
		PubServiceKlineInfoLastProcedure,
		svc.KlineInfoLast,
		connect.WithSchema(pubServiceKlineInfoLastMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceKlineTradeListHandler := connect.NewUnaryHandler(
		PubServiceKlineTradeListProcedure,
		svc.KlineTradeList,
		connect.WithSchema(pubServiceKlineTradeListMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceRankListHandler := connect.NewUnaryHandler(
		PubServiceRankListProcedure,
		svc.RankList,
		connect.WithSchema(pubServiceRankListMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceSendSmsHandler := connect.NewUnaryHandler(
		PubServiceSendSmsProcedure,
		svc.SendSms,
		connect.WithSchema(pubServiceSendSmsMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	pubServiceGreetHandler := connect.NewServerStreamHandler(
		PubServiceGreetProcedure,
		svc.Greet,
		connect.WithSchema(pubServiceGreetMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/mproto.PubService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case PubServiceIndexProcedure:
			pubServiceIndexHandler.ServeHTTP(w, r)
		case PubServiceHelpListProcedure:
			pubServiceHelpListHandler.ServeHTTP(w, r)
		case PubServiceHelpDetailProcedure:
			pubServiceHelpDetailHandler.ServeHTTP(w, r)
		case PubServiceCarouselListProcedure:
			pubServiceCarouselListHandler.ServeHTTP(w, r)
		case PubServiceIndexAllCoinProcedure:
			pubServiceIndexAllCoinHandler.ServeHTTP(w, r)
		case PubServiceBankListProcedure:
			pubServiceBankListHandler.ServeHTTP(w, r)
		case PubServiceCoinManagementListProcedure:
			pubServiceCoinManagementListHandler.ServeHTTP(w, r)
		case PubServiceMinerListProcedure:
			pubServiceMinerListHandler.ServeHTTP(w, r)
		case PubServiceGetBankByPreProcedure:
			pubServiceGetBankByPreHandler.ServeHTTP(w, r)
		case PubServiceKlineInfoListProcedure:
			pubServiceKlineInfoListHandler.ServeHTTP(w, r)
		case PubServiceKlineInfoLastProcedure:
			pubServiceKlineInfoLastHandler.ServeHTTP(w, r)
		case PubServiceKlineTradeListProcedure:
			pubServiceKlineTradeListHandler.ServeHTTP(w, r)
		case PubServiceRankListProcedure:
			pubServiceRankListHandler.ServeHTTP(w, r)
		case PubServiceSendSmsProcedure:
			pubServiceSendSmsHandler.ServeHTTP(w, r)
		case PubServiceGreetProcedure:
			pubServiceGreetHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedPubServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedPubServiceHandler struct{}

func (UnimplementedPubServiceHandler) Index(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.IndexResonse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.Index is not implemented"))
}

func (UnimplementedPubServiceHandler) HelpList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.HelpListResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.HelpList is not implemented"))
}

func (UnimplementedPubServiceHandler) HelpDetail(context.Context, *connect.Request[mproto.StringParam]) (*connect.Response[model.Config], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.HelpDetail is not implemented"))
}

func (UnimplementedPubServiceHandler) CarouselList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.CarouselListResonse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.CarouselList is not implemented"))
}

func (UnimplementedPubServiceHandler) IndexAllCoin(context.Context, *connect.Request[mproto.PidParam]) (*connect.Response[mproto.IndexAllCoinResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.IndexAllCoin is not implemented"))
}

func (UnimplementedPubServiceHandler) BankList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.BankListResonse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.BankList is not implemented"))
}

func (UnimplementedPubServiceHandler) CoinManagementList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.CoinManagementListResonses], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.CoinManagementList is not implemented"))
}

func (UnimplementedPubServiceHandler) MinerList(context.Context, *connect.Request[mproto.NullMsg]) (*connect.Response[mproto.MinerListResonses], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.MinerList is not implemented"))
}

func (UnimplementedPubServiceHandler) GetBankByPre(context.Context, *connect.Request[mproto.GetBankByPreRequest]) (*connect.Response[model.FinancialBank], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.GetBankByPre is not implemented"))
}

func (UnimplementedPubServiceHandler) KlineInfoList(context.Context, *connect.Request[mproto.KlineInfoRequest]) (*connect.Response[mproto.KlineDataResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.KlineInfoList is not implemented"))
}

func (UnimplementedPubServiceHandler) KlineInfoLast(context.Context, *connect.Request[mproto.KlineInfoRequest]) (*connect.Response[mproto.KlineInfoLastResonse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.KlineInfoLast is not implemented"))
}

func (UnimplementedPubServiceHandler) KlineTradeList(context.Context, *connect.Request[mproto.StringParam]) (*connect.Response[mproto.KlineTradeListResonse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.KlineTradeList is not implemented"))
}

func (UnimplementedPubServiceHandler) RankList(context.Context, *connect.Request[mproto.RankListRequest]) (*connect.Response[mproto.RankListResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.RankList is not implemented"))
}

func (UnimplementedPubServiceHandler) SendSms(context.Context, *connect.Request[mproto.SendSmsRequest]) (*connect.Response[mproto.MsgResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.SendSms is not implemented"))
}

func (UnimplementedPubServiceHandler) Greet(context.Context, *connect.Request[mproto.StringParam], *connect.ServerStream[mproto.GreetResponse]) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("mproto.PubService.Greet is not implemented"))
}
