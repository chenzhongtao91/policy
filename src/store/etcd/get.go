package etcd

import (
	"fmt"
	"strconv"

	"github.com/coreos/etcd/client"
)

/*
func (estore *EtcdStoreDriver) Get(key string, sorted bool, quorum bool) (*client.Response, error) {
	ctx, cancel := contextWithTotalTimeout()
	resp, err := estore.keysApi.Get(ctx, key, &client.GetOptions{Sort: sorted, Quorum: quorum})
	cancel()
	if err != nil {
		return resp, err
	}

	return resp, nil
}
*/

func (estore *EtcdStoreDriver) Get(key string, opts map[string]string) (string, error) {
	sorted := false
	quorum := false
	var err error

	value, ok := opts[OPT_GET_SORTED]
	if ok {
		sorted, err = strconv.ParseBool(value)
		if err != nil {
			return "", err
		}
	}

	value, ok = opts[OPT_GET_QUORUM]
	if ok {
		quorum, err = strconv.ParseBool(value)
		if err != nil {
			return "", err
		}
	}

	ctx, cancel := contextWithTotalTimeout()

	options := &client.GetOptions{
		Sort:   sorted,
		Quorum: quorum,
	}

	resp, err := estore.keysApi.Get(ctx, key, options)
	cancel()

	fmt.Println("store.List", err)
	if err != nil {
		if cerr, ok := err.(*client.ClusterError); ok {
			fmt.Println(cerr)
			for i, _ := range cerr.Errors {
				if i == 100 {
					return "", nil
				} else {
					return "", err
				}
			}
		}
		return "", err
	}

	if resp.Node.Dir {
		return "", fmt.Errorf("directory")
	}

	return resp.Node.Value, nil
}
