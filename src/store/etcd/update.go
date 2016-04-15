package etcd

import (
	"time"

	"github.com/coreos/etcd/client"
)

func (estore *EtcdStoreDriver) Update(key string, value string, ttl int) (*client.Response, error) {
	ctx, cancel := contextWithTotalTimeout()
	resp, err := estore.keysApi.Set(ctx, key, value, &client.SetOptions{TTL: time.Duration(ttl) * time.Second, PrevExist: client.PrevExist})
	cancel()
	if err != nil {
		return nil, err
	}

	return resp, err
}
