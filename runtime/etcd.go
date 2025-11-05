package runtime

import (
	"context"
	"time"

	"go.etcd.io/etcd/client/v3"
)

type EtcdClient struct {
	cli          *clientv3.Client
	readTimeout  time.Duration
	writeTimeout time.Duration
}

func NewEtcdClient(config *EtcdConfig) (*EtcdClient, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Endpoints,
		DialTimeout: time.Duration(config.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return &EtcdClient{
		cli:          cli,
		readTimeout:  time.Duration(config.ReadTimeout) * time.Second,
		writeTimeout: time.Duration(config.WriteTimeout) * time.Second,
	}, nil
}

func (e *EtcdClient) Put(key, val string) (*clientv3.PutResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.writeTimeout)
	defer cancel()

	return e.cli.Put(ctx, key, val)
}

func (e *EtcdClient) Get(key string) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.readTimeout)
	defer cancel()

	return e.cli.Get(ctx, key)
}

func (e *EtcdClient) GetCountWithPrefix(prefix string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.readTimeout)
	defer cancel()

	resp, err := e.cli.Get(ctx, prefix, clientv3.WithCountOnly(), clientv3.WithPrefix())
	if err != nil {
		return 0, err
	}

	return resp.Count, nil
}

func (e *EtcdClient) RangeWithPrefix(prefix, from string, limit int64) (*clientv3.GetResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), e.readTimeout)
	defer cancel()

	endKey := clientv3.GetPrefixRangeEnd(prefix)

	return e.cli.Get(
		ctx,
		prefix+from,
		clientv3.WithRange(endKey),
		clientv3.WithLimit(limit),
		clientv3.WithSort(clientv3.SortByValue, clientv3.SortAscend),
	)
}

func (e *EtcdClient) Close() error {
	return e.cli.Close()
}
