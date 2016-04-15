package etcd

import (
	"github.com/coreos/etcd/client"
)

func (estore *EtcdStoreDriver) RmDir(key string) (*client.Response, error) {
	ctx, cancel := contextWithTotalTimeout()

	resp, err := estore.keysApi.Delete(ctx, key, &client.DeleteOptions{Dir: true})

	cancel()
	if err != nil {
		return nil, err
	}

	return resp, nil
}
