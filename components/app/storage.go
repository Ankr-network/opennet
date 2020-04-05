package app

import (
	"time"

	"github.com/coreos/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"
)

const (
	keyPath    = "/ssl/etcd-key.pem"
	certPath   = "/ssl/etcd.pem"
	caCertPath = "/ssl/etcd-root-ca.pem"
	// keyPath    = "/home/mobius/ssl/etcd-key.pem"
	// certPath   = "/home/mobius/ssl/etcd.pem"
	// caCertPath = "/home/mobius/ssl/etcd-root-ca.pem"
)

var (
	// Endpoints etcd
	Endpoints []string
	store     *client
)

// client storage ops client
type client struct {
	c *clientv3.Client
}

// InitStore ...
func InitStore() error {
	tlsInfo := transport.TLSInfo{
		CertFile:      certPath,
		KeyFile:       keyPath,
		TrustedCAFile: caCertPath,
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return err
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   Endpoints,
		DialTimeout: 30 * time.Second,
		TLS:         tlsConfig,
	})
	if err != nil {
		return err
	}
	store = &client{c: cli}
	return nil
}
