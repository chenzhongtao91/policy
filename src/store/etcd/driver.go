package etcd

import (
	"dmutex"
	"store"

	"github.com/coreos/etcd/client"
)

const (
	OPT_GET_SORTED = "sorted"
	OPT_GET_QUORUM = "qurum"

	OPT_SET_TTL       = "ttl"
	OPT_SET_PREVVALUE = "prevValue"
	OPT_SET_PREVINDEX = "prevIndex"

	OPT_LIST_RECURSIVE = "recursive"
	OPT_LIST_SORTED    = "sorted"
	OPT_LIST_QUORUM    = "quorum"

	OPT_REMOVE_RECURSIVE = "recursive"
	OPT_REMOVE_DIR       = "dir"
	OPT_REMOVE_PREVVALUE = "prevValue"
	OPT_REMOVE_PREVINDEX = "prevIndex"
)

type EtcdStoreDriver struct {
	keysApi    client.KeysAPI
	storeMutex *dmutex.Mutex
}

const (
	LockTimeOut = 10
	EtcdUrls    = "http://127.0.0.1:2379"
)

func NewStore() {
	if store.Backend != nil {
		return
	}

	estore := new(EtcdStoreDriver)
	estore.keysApi = mustNewKeyAPI()
	estore.storeMutex = dmutex.NewMutex(store.STORELOCK, LockTimeOut, []string{EtcdUrls})
	if estore.storeMutex == nil {
		panic("Store Invalid")
	}

	store.Backend = estore
}

func (estore *EtcdStoreDriver) Lock() error {
	return estore.storeMutex.Lock()
}

func (estore *EtcdStoreDriver) Unlock() error {
	return estore.storeMutex.Unlock()
}
