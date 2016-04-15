package scheduler

import (
	"fmt"
	"sort"
	"strconv"

	"meta"
	"meta/proto"
)

const (
	FilterCapacity = "FilterCapacity"

	WeigherCapacity = "WeigherCapacity"

	Backend = "Backend"
	Replica = "Replica"
)

func DoScheduler(opts map[string]string) ([]*metaproto.Device, error) {

	backend, ok := opts[Backend]
	if !ok {
		return nil, fmt.Errorf("ERROR: the key '%s' is required", Backend)
	}

	_, ok = opts[FilterCapacity]
	if !ok {
		return nil, fmt.Errorf("ERROR: the key 's' is required", FilterCapacity)
	}

	Rep, ok := opts[Replica]
	if !ok {
		return nil, fmt.Errorf("ERROR: the key '%s' is required", Replica)
	}
	replica, _ := strconv.ParseInt(Rep, 10, 64)

	fmt.Println(backend)

	devices, err := metadata.GetFreeDevices(backend)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v\n", devices)

	/*
		devices := []*metaproto.Device{
			&metaproto.Device{Id: []byte("1"), Total: []byte("200")},
			&metaproto.Device{Id: []byte("1"), Total: []byte("400")},
			&metaproto.Device{Id: []byte("1"), Total: []byte("600")}}
	*/

	filters, _ := MakeFilters(opts)

	for _, filter := range filters {
		devices = filter.Filter(devices)
	}

	if len(devices) == 0 {
		fmt.Println("ERROR: not device is matched")
		return nil, fmt.Errorf("ERROR: not device is matched")
	}
	if len(devices) < int(replica) {
		return nil, fmt.Errorf("ERROR: the device number is less replica number")
	}

	weighers, _ := MakeWeighers(opts)
	if len(weighers) == 0 {
		opts[WeigherCapacity] = "100"
		weighers, _ = MakeWeighers(opts)
	}

	allCost := make([]float64, len(devices))
	for _, weigher := range weighers {
		allCost, _ = SliceAddSliceFloat64(allCost, weigher.Weigher(devices))
	}
	sort.Sort(DeviceCostWrapper{devices, allCost})
	fmt.Printf("%v\n", devices)
	return devices[0:replica], nil
}
