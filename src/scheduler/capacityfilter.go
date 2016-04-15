package scheduler

import (
	"strconv"

	"meta/proto"
)

type CapacityFilter struct {
	Capacity int64
}

func (cfilter *CapacityFilter) Filter(devices []*metaproto.Device) []*metaproto.Device {
	deviceFilter := []*metaproto.Device{}
	for _, v := range devices {
		vTotal, _ := strconv.ParseInt(string(v.Total), 10, 64)
		if vTotal >= cfilter.Capacity {
			deviceFilter = append(deviceFilter, v)
		}
	}
	//sort.Sort(DeviceWrapper{deviceFilter, ByTotal})
	return deviceFilter
}

func MakeCapacityFilter(value string) (Filter, error) {
	capacity, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, err
	}
	var filterPtr Filter = &CapacityFilter{capacity}
	return filterPtr, nil

}
