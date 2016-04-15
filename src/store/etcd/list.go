package etcd

import (
	"strconv"

	"github.com/coreos/etcd/client"
)

/*
func (estore *EtcdStoreDriver) List(key string, recursive bool, sort bool, quorum bool) (*client.Response, error) {
	tkey := "/"
	if len(key) != 0 {
		tkey = key
	}

	ctx, cancel := contextWithTotalTimeout()
	resp, err := estore.keysApi.Get(ctx, tkey, &client.GetOptions{Sort: sort, Recursive: recursive, Quorum: quorum})
	cancel()
	if err != nil {
		return nil, err
	}

	return resp, nil
}
*/

func (estore *EtcdStoreDriver) List(key string, opts map[string]string) ([]string, error) {
	recursive := false
	sort := false
	quorum := false
	var err error
	value, ok := opts[OPT_LIST_RECURSIVE]
	if ok {
		recursive, err = strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
	}

	value, ok = opts[OPT_LIST_SORTED]
	if ok {
		sort, err = strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
	}

	value, ok = opts[OPT_LIST_QUORUM]
	if ok {
		quorum, err = strconv.ParseBool(value)
		if err != nil {
			return nil, err
		}
	}

	tkey := "/"
	if len(key) != 0 {
		tkey = key
	}

	ctx, cancel := contextWithTotalTimeout()
	resp, err := estore.keysApi.Get(ctx, tkey, &client.GetOptions{Sort: sort, Recursive: recursive, Quorum: quorum})
	cancel()
	if err != nil {
		return nil, err
	}

	values := []string{}
	for _, node := range resp.Node.Nodes {
		if !node.Dir {
			values = append(values, node.Key)
		}
	}

	return values, nil
}
