package gateway

import (
	"context"
	"encoding/json"
	"github.com/skip2/go-qrcode"
	"google.golang.org/grpc"
	"log"
	"muskex/gen/mproto"
	"muskex/gen/mproto/model"
	"muskex/rserver"
	"muskex/rserver/ws"
	"muskex/utils"
	"net/http"

	"github.com/golang/glog"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

// Endpoint describes a gRPC endpoint
type Endpoint struct {
	Network, Addr string
}

// Options is a set of options to be passed to Run
type Options struct {
	// Addr is the address to listen
	Addr string

	// GRPCServer defines an endpoint of a gRPC service
	GRPCServer Endpoint

	// OpenAPIDir is a path to a directory from which the server
	// serves OpenAPI specs.
	OpenAPIDir string

	// Mux is a list of options to be passed to the gRPC-Gateway multiplexer
	Mux []gwruntime.ServeMuxOption
}

// Run starts a HTTP server and blocks while running if successful.
// The server will be shutdown when "ctx" is canceled.
func Run(ctx context.Context, opts Options, gwRegs []func(context.Context, *gwruntime.ServeMux, *grpc.ClientConn) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	conn, err := dial(ctx, opts.GRPCServer.Network, opts.GRPCServer.Addr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		if err := conn.Close(); err != nil {
			glog.Errorf("Failed to close a client connection to the gRPC server: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/openapiv2/", openAPIServer(opts.OpenAPIDir))
	mux.HandleFunc("/healthz", healthzServer(conn))
	mux.HandleFunc("/face_verify_callback", func(w http.ResponseWriter, req *http.Request) {
		log.Println("face_verify_callback", req.URL.String())
		certId := req.URL.Query().Get("certifyId")
		passed := req.URL.Query().Get("passed")
		verifyed := int64(3)
		if passed == "200" {
			verifyed = 2
		}
		utils.Orm.Where(model.User{CertifyId: certId}).Updates(&model.User{CertifyId: certId, IsCertified: verifyed})
		w.Write([]byte("{\"message\":ok}"))
		return
	})
	mux.HandleFunc("/qrcode", func(w http.ResponseWriter, req *http.Request) {
		content := req.URL.Query().Get("content")
		png, err := qrcode.Encode(content, qrcode.Medium, 256)
		if err != nil {
			log.Println("qrcode err", err)
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "max-age=800000")
		w.Write(png)
		return
	})
	mux.HandleFunc("/u_trade_callback", rserver.TradeCallbackHandler)
	mux.HandleFunc("/gcfg", func(w http.ResponseWriter, request *http.Request) {
		bs, _ := json.Marshal(mproto.GCfg)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(bs)
	})
	ws.InitTradeHub(rserver.KlineIdMap)
	mux.HandleFunc("/ws", ws.WsHandle)

	gw, err := newGateway(ctx, conn, opts.Mux, gwRegs)
	if err != nil {
		return err
	}
	prettier := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("Xhost", r.Host)
			h.ServeHTTP(w, r)
		})
	}
	mux.Handle("/", prettier(gw))

	s := &http.Server{
		Addr:    opts.Addr,
		Handler: allowCORS(mux),
	}
	go func() {
		<-ctx.Done()
		glog.Infof("Shutting down the http server")
		if err := s.Shutdown(context.Background()); err != nil {
			glog.Errorf("Failed to shutdown http server: %v", err)
		}
	}()

	glog.Infof("Starting listening at %s", opts.Addr)
	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		glog.Errorf("Failed to listen and serve: %v", err)
		return err
	}
	return nil
}
