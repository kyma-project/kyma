package etcd

import "github.com/coreos/etcd/clientv3"

// Client wraps etcd client for testing purposes.
type Client interface {
	clientv3.KV
}

// NewClient produces new, configured etcd client.
func NewClient(cfg Config) (Client, error) {

	etcdCfg := clientv3.Config{
		Endpoints: cfg.Endpoints,
		Username:  cfg.Username,
		Password:  cfg.Password,
	}

	cli, err := clientv3.New(etcdCfg)

	return cli, err
}
