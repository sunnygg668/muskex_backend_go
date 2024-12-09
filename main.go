package main

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	runtime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/spf13/pflag"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"log"
	"muskex/gateway"
	"muskex/gen/mproto"
	"muskex/gen/mproto/model"
	"muskex/gen/mproto/mprotoconnect"
	"muskex/rserver"
	"muskex/utils"
	"muskex/utils/signal"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)
import (
	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"
)

// withCORS adds CORS support to a Connect HTTP handler.
func withCORS(connectHandler http.Handler) http.Handler {
	c := cors.New(cors.Options{
		//AllowedOrigins: []string{"http://localhost:5173"}, // replace with your domain
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: append(connectcors.AllowedHeaders(), "batoken"),
		ExposedHeaders: connectcors.ExposedHeaders(),
		MaxAge:         7200, // 2 hours in seconds
	})
	return c.Handler(connectHandler)
}

const TokenHeader = "token"

func main() {
	log.SetFlags(log.LstdFlags)
	log.SetOutput(os.Stdout)
	var dbUrl, serverPort, gServerPort, env, pid, pubDomain string
	var onlyKline, printFlags bool
	pflag.StringVarP(&dbUrl, "db", "d", "admin:UhwZR0BApOSD57qATCao@tcp(musk-ex2-2024-11-155.cuyych8yxu1j.ap-southeast-1.rds.amazonaws.com:3306)/saas?loc=Local&parseTime=true&multiStatements=true", "mysql database url")
	pflag.StringVarP(&gServerPort, "gport", "g", "8079", "grpc api service port")
	pflag.StringVarP(&serverPort, "port", "p", "8080", "restful　service port")
	pflag.StringVarP(&env, "env", "e", "dev", "环境名字debug prod test")
	pflag.StringVarP(&pid, "pid", "", "1", "平台id: 1 2 3 4")
	pflag.StringVarP(&pubDomain, "pub_domain", "", "http://127.0.0.1", "平台域名地址")
	//pflag.BoolVarP(&sync_price, "sync_price", "", false, "同步ticker价格到数据库")
	pflag.BoolVarP(&onlyKline, "only_kline", "", true, "只提供kline api")
	pflag.BoolVarP(&printFlags, "print", "", false, "")
	pflag.Parse()

	mproto.GCfg.PubDomain = pubDomain
	mproto.GCfg.Env = env
	mproto.GCfg.Pid = pid
	if printFlags {
		log.Println("db:", dbUrl)
		log.Printf("gcfg:%#v", mproto.GCfg)
		os.Exit(0)
	}

	shutdownChan, err := signal.Intercept()
	if err != nil {
		panic(err)
	}
	utils.InitDb(dbUrl)

	if onlyKline {
		rserver.InitKlineData()
		go rserver.SubData(env == "dev")
		go rserver.SyncCoinPrice()
	} else {
		utils.Orm.AutoMigrate(model.TokenV2{})
		cfg := rserver.LoadDbConfig().NameValues
		coinCode, _ := strconv.Atoi(cfg["udun_main_coin_code"])
		rserver.InitializeCreateReqConfig(cfg["udun_api_key"], cfg["udun_gateway_address"]+"/mch/address/create", cfg["udun_merchant_no"], coinCode, pubDomain+"/udun_trade_callback?pid="+pid)
	}

	mux := http.NewServeMux()
	interceptors := connect.WithInterceptors(NewAuthInterceptor(""))
	mux.Handle(mprotoconnect.NewPubServiceHandler(rserver.PubServer{}, interceptors))
	if !onlyKline {
		mux.Handle(mprotoconnect.NewUserAuthServiceHandler(rserver.UserAuthServer{}, interceptors))
		mux.Handle(mprotoconnect.NewWpAuthServiceHandler(rserver.LoginServer{}, interceptors))
		mux.Handle(mprotoconnect.NewUserManServiceHandler(rserver.UserManServer{}, interceptors))
	}

	go func() {
		log.Println("grpc server start:", "0.0.0.0:"+gServerPort)
		err := http.ListenAndServe(
			"0.0.0.0:"+gServerPort,
			// Use h2c so we can serve HTTP/2 without TLS.
			withCORS(h2c.NewHandler(mux, &http2.Server{})),
		)
		log.Println("grpc server end:", err)
	}()

	customHeaderMatcher := func(key string) (string, bool) {
		//log.Println("header key", key)
		switch key {
		case "Token":
			return key, true
		case "Host", "Xhost", "Pid", "pid":
			return key, true
		default:
			return runtime.DefaultHeaderMatcher(key)
		}
	}
	handleRoutingError := func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, httpStatus int) {
		log.Println("RoutingError: path ", r.URL.Path, "method", r.Method)
		runtime.DefaultRoutingErrorHandler(ctx, mux, marshaler, w, r, httpStatus)
	}
	//gw server
	opts := gateway.Options{
		Addr: ":" + serverPort,
		GRPCServer: gateway.Endpoint{
			Network: "tcp",
			Addr:    "0.0.0.0:" + gServerPort,
		},
		OpenAPIDir: "swagger",
		Mux: []runtime.ServeMuxOption{
			runtime.WithRoutingErrorHandler(handleRoutingError),
			runtime.WithIncomingHeaderMatcher(customHeaderMatcher)},
	}
	go func() {
		gwRegs := []func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error{}
		gwRegs = append(gwRegs, mproto.RegisterPubServiceHandler)
		if !onlyKline {
			gwRegs = append(gwRegs, mproto.RegisterUserAuthServiceHandler)
			gwRegs = append(gwRegs, mproto.RegisterWpAuthServiceHandler)
			gwRegs = append(gwRegs, mproto.RegisterUserManServiceHandler)
			utils.InitCapcha()
		}
		gateway.Run(context.Background(), opts, gwRegs)
	}()
	log.Println("http server start:", "0.0.0.0:"+serverPort)
	<-shutdownChan.ShutdownChannel()
	time.Sleep(1 * time.Second)
	log.Println("done shutdown")
}

func NewAuthInterceptor(clientToken string) connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			log.Println("host:", req.Header().Get("Xhost"), "path:", req.Spec().Procedure, "method", req.HTTPMethod())
			//log.Println("header ", req.Header())

			//host := req.Header().Get("Xhost")
			//ctx = context.WithValue(ctx, "pid", utils.DomainPidMap[host])

			path := req.Spec().Procedure
			if strings.HasPrefix(path, "/mproto.PubService") {
				return next(ctx, req)
			}
			if strings.HasPrefix(path, "/mproto.UserAuthService") {
				return next(ctx, req)
			}
			if req.Spec().IsClient {
				// Send a token with client requests.
				req.Header().Set(TokenHeader, clientToken)
				return next(ctx, req)
			} else if req.Header().Get(TokenHeader) == "" {
				//Check token in handlers.
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("请先登陆"),
				)
			}
			tokenstr := req.Header().Get(TokenHeader)
			if tokenstr != "" {
				userId, err := getAuthedUserId(tokenstr)
				if err != nil {
					return nil, err
				}
				ctx = context.WithValue(ctx, "userId", userId)
			}
			return next(ctx, req)
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
func getAuthedUserId(tokenstr string) (int64, error) {
	//get := func() (any, error) {
	//	hash1 := hmac.New(ripemd160.New, []byte("DS3xQhvAoJnWUGkcLfCP8uswi4YqbIdr"))
	//	hash1.Write([]byte(tokenstr))
	//	tokenCode := fmt.Sprintf("%x", hash1.Sum(nil))
	//	token := &model.Token{Token: tokenCode}
	//	err := utils.Orm.First(token, token).Error
	//	if err != nil {
	//		return nil, connect.NewError(
	//			connect.CodeUnauthenticated,
	//			errors.New("err token"))
	//	}
	//	if token.ExpireTime < time.Now().Unix() {
	//		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("token expired"))
	//	}
	//	return token.UserId, nil
	//}
	get := func() (any, error) {
		token := &model.TokenV2{Token: tokenstr}
		err := utils.Orm.First(token, token).Error
		if err != nil && err == gorm.ErrRecordNotFound {
			return 0, status.Error(codes.Unauthenticated, "token not found")
		} else if err != nil {
			return 0, err
		} else {
			if token.Disabled {
				return 0, status.Error(codes.Unauthenticated, "token disabled")
			} else if time.Now().Sub(token.UpdatedAt).Hours() > 48 {
				return 0, status.Error(codes.Unauthenticated, "登陆已经过期，请重新登陆")
			} else {
				return token.UserId, nil
			}
		}
		return 0, nil
	}
	res, err := utils.CacheFromLru(0, "token_userid"+tokenstr, 7200, get)
	if err != nil {
		return 0, err
	}
	return res.(int64), nil
}
