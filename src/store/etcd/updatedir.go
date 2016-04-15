package etcd

import (
	"time"

	"github.com/coreos/etcd/client"
)

func (estore *EtcdStoreDriver) UpdateDir(key string, ttl int) (*client.Response, error) {
	ctx, cancel := contextWithTotalTimeout()

	resp, err := estore.keysApi.Set(ctx, key, "", &client.SetOptions{TTL: time.Duration(ttl) * time.Second, Dir: true, PrevExist: client.PrevExist})
	cancel()
	if err != nil {
		return nil, err
	}

	return resp, nil
}
