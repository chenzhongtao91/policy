package etcd

import (
	"strconv"
	"time"

	"github.com/coreos/etcd/client"
)

/*
 *  Response in file etcd/client/keys.go
 */
func (estore *EtcdStoreDriver) _set(key string, value string, ttl int64, prevValue string, prevIndex int64) (*client.Response, error) {
	ctx, cancel := contextWithTotalTimeout()
	resp, err := estore.keysApi.Set(ctx, key, value, &client.SetOptions{TTL: time.Duration(ttl) * time.Second,
		PrevIndex: uint64(prevIndex), PrevValue: prevValue})
	cancel()
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func (estore *EtcdStoreDriver) Set(key string, val string, opts map[string]string) error {
	ttl := int64(0)
	prevValue := ""
	prevIndex := int64(0)
	var err error

	value, ok := opts[OPT_SET_TTL]
	if ok {
		ttl, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
	}

	value, ok = opts[OPT_SET_PREVVALUE]
	if ok {
		prevValue = value
	}

	value, ok = opts[OPT_SET_PREVINDEX]
	if ok {
		prevIndex, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
	}

	_, err = estore._set(key, val, ttl, prevValue, prevIndex)
	if err != nil {
		return err
	}

	return nil
}
