package main

import (
	"fmt"

	"meta/proto"
	"scheduler"
	"store/etcd"
)

func test_filter() {
	fmt.Printf("test filter:\n")
	devices := []*metaproto.Device{
		&metaproto.Device{Id: []byte("1"), Total: []byte("200")},
		&metaproto.Device{Id: []byte("2"), Total: []byte("400")},
		&metaproto.Device{Id: []byte("3"), Total: []byte("600")}}

	filter1 := scheduler.CapacityFilter{Capacity: 100}
	deviceFilter := filter1.Filter(devices)
	fmt.Printf("%v\n", deviceFilter)
}

func test_scheduler() {
	fmt.Printf("\ntest scheduler:\n")
	var opts = map[string]string{"FilterCapacity": "200", "WeigherCapacity": "100", "Backend": "ceph", "Replica": "1"}
	deviceptr, err := scheduler.DoScheduler(opts)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%v\n", deviceptr)

}

func main() {
	etcd.NewStore()
	test_filter()
	test_scheduler()
}
