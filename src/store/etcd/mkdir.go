package etcd

import (
	"time"

	"github.com/coreos/etcd/client"
)

func (estore *EtcdStoreDriver) MkDir(key string, prevExist client.PrevExistType, ttl int) (*client.Response, error) {
	ctx, cancel := contextWithTotalTimeout()
	resp, err := estore.keysApi.Set(ctx, key, "", &client.SetOptions{TTL: time.Duration(ttl) * time.Second, Dir: true, PrevExist: prevExist})
	cancel()
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (estore *EtcdStoreDriver) SetDir(key string, ttl int) (*client.Response, error) {
	return estore.MkDir(key, client.PrevIgnore, ttl)
}
