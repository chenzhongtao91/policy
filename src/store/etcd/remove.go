package etcd

import (
	"strconv"

	"github.com/coreos/etcd/client"
)

/*
func (estore *EtcdStoreDriver) Remove(key string, recursive bool, dir bool, prevValue string, prevIndex int) (*client.Response, error) {
	ctx, cancel := contextWithTotalTimeout()
	resp, err := estore.keysApi.Delete(ctx, key, &client.DeleteOptions{PrevIndex: uint64(prevIndex), PrevValue: prevValue, Dir: dir, Recursive: recursive})
	cancel()
	if err != nil {
		return resp, err
	}

	return resp, nil
}
*/

func (estore *EtcdStoreDriver) Remove(key string, opts map[string]string) error {
	recursive := false
	dir := false
	prevValue := ""
	prevIndex := int64(0)
	var err error

	value, ok := opts[OPT_REMOVE_RECURSIVE]
	if ok {
		recursive, err = strconv.ParseBool(value)
		if err != nil {
			return err
		}
	}

	value, ok = opts[OPT_REMOVE_DIR]
	if ok {
		dir, err = strconv.ParseBool(value)
		if err != nil {
			return err
		}
	}

	value, ok = opts[OPT_SET_PREVVALUE]
	if ok {
		prevValue = value
	}

	value, ok = opts[OPT_REMOVE_PREVINDEX]
	if ok {
		prevIndex, err = strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
	}

	ctx, cancel := contextWithTotalTimeout()
	_, err = estore.keysApi.Delete(ctx, key, &client.DeleteOptions{PrevIndex: uint64(prevIndex), PrevValue: prevValue, Dir: dir, Recursive: recursive})
	cancel()
	if err != nil {
		return err
	}

	return nil
}
