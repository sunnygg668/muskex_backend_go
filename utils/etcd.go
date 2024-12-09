package utils
import(
	"crypto/tls"
	"crypto/x509"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"io/ioutil"
	"log"
	"strings"
	"time"
)

var EtcdCli *clientv3.Client
func InitEtcd(endpoins string) {
	cfg := clientv3.Config{
		Endpoints:strings.Split(endpoins,","),
		DialTimeout: 5* time.Second,
	}
	tlsConfig,err:=guestEtcdClientTlsConfig()
	if err != nil {
		log.Fatalln(err)
	}
	cfg.TLS=tlsConfig
	for{
		cli, err := clientv3.New(cfg)
		if err != nil {
			log.Println("etcd connect",err)
			continue
		}
		EtcdCli=cli
		break
	}
	log.Println("etcd ok")
}
func guestEtcdClientTlsConfig()(*tls.Config,error){
	pemServerCA, err := ioutil.ReadFile("asset/ca-cert.pem")
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}
	// Load client's certificate and private key
	clientCert, err := tls.LoadX509KeyPair("asset/client-cert.pem", "asset/client-key.pem")
	if err != nil {
		return nil, err
	}
	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
	}
	return config,nil
}