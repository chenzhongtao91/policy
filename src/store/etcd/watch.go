package etcd

/*
import (
	"fmt"
	"os"
	"os/signal"

	"golang.org/x/net/context"
	"github.com/coreos/etcd/client"
)


func (estore *EtcdStoreDriver) Watch(key string, recursive bool, forever bool, index int) (*client.Response, error) {
	ki := mustNewKeyAPI()

	stop := false
	w := ki.Watcher(key, &client.WatcherOptions{AfterIndex: uint64(index), Recursive: recursive})

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)

	go func() {
		<-sigch
		os.Exit(0)
	}()

	for !stop {
		resp, err := w.Next(context.TODO())
		if err != nil {
			return nil, err
		}
		if resp.Node.Dir {
			continue
		}
		if recursive {
			fmt.Printf("[%s] %s\n", resp.Action, resp.Node.Key)
		}

		//callback

		if !forever {
			stop = true
		}
	}
}
*/
