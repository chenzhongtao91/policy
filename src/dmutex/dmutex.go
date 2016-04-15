package dmutex

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
)

const (
	defaultTTL   = 10
	defaultRetry = 3
	deleteAction = "delete"
	expireAction = "expire"
)

type Mutex struct {
	key    string
	id     string
	client client.Client
	kapi   client.KeysAPI
	ctx    context.Context
	ttl    time.Duration
	mutex  *sync.Mutex
}

func NewMutex(key string, ttl int, hosts []string) *Mutex {
	cfg := client.Config{
		Endpoints:               hosts,
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(cfg)
	if err != nil {
		return nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil
	}

	if len(key) == 0 || len(hosts) == 0 {
		return nil
	}

	if key[0] != '/' {
		key = "/" + key
	}

	if ttl < 1 {
		ttl = defaultTTL
	}

	return &Mutex{
		key:    key,
		id:     fmt.Sprintf("%v-%v-%v", hostname, os.Getpid(), time.Now().Format("20160327-13:37:21.999999999")),
		client: c,
		kapi:   client.NewKeysAPI(c),
		ctx:    context.TODO(),
		ttl:    time.Second * time.Duration(ttl),
		mutex:  new(sync.Mutex),
	}
}

func (mutex *Mutex) Lock() (err error) {
	mutex.mutex.Lock()
	for try := 1; try <= defaultRetry; try++ {
		err = mutex.lock()
		if err == nil {
			return nil
		}

		logrus.Debugf("Lock %v ERROR %v", mutex.key, err)
		if try < defaultRetry {
			logrus.Debugf("Retry to lock %v again", mutex.key)
		}
	}

	return nil
}

func (mutex *Mutex) lock() (err error) {
	setOptions := &client.SetOptions{
		PrevExist: client.PrevNoExist,
		TTL:       mutex.ttl,
	}

	for {
		resp, err := mutex.kapi.Set(mutex.ctx, mutex.key, mutex.id, setOptions)
		if err == nil {
			logrus.Debugf("Create node %v OK [%q]", mutex.key, resp)
			return nil
		}

		e, ok := err.(client.Error)
		if !ok {
			return err
		}

		if e.Code != client.ErrorCodeNodeExist {
			return err
		}

		resp, err = mutex.kapi.Get(mutex.ctx, mutex.key, nil)
		if err != nil {
			return err
		}

		logrus.Debugf("Get Key %v OK", mutex.key)
		watcherOptions := &client.WatcherOptions{
			AfterIndex: resp.Index,
			Recursive:  false,
		}

		watcher := mutex.kapi.Watcher(mutex.key, watcherOptions)
		for {
			logrus.Debugf("Watching %v ...", mutex.key)
			resp, err = watcher.Next(mutex.ctx)
			if err != nil {
				return err
			}

			logrus.Debugf("Received an event: %q", resp)
			if resp.Action == deleteAction || resp.Action == expireAction {
				break
			}
		}
	}
	return err
}

func (mutex *Mutex) Unlock() (err error) {
	defer mutex.mutex.Unlock()

	for i := 1; i <= defaultRetry; i++ {
		var resp *client.Response
		resp, err = mutex.kapi.Delete(mutex.ctx, mutex.key, nil)
		if err == nil {
			logrus.Debugf("Delete %v OK", mutex.key)
			return nil
		}
		logrus.Debugf("Delete %v failed: %q", mutex.key, resp)

		e, ok := err.(client.Error)
		if ok && e.Code == client.ErrorCodeKeyNotFound {
			return nil
		}
	}
	return err
}
