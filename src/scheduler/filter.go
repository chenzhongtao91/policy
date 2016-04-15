package scheduler

import (
	"meta/proto"
)

type Filter interface {
	Filter([]*metaproto.Device) []*metaproto.Device
}

type makeFilterFunc func(string) (Filter, error)

var FilterFactory = map[string]makeFilterFunc{
	FilterCapacity: MakeCapacityFilter,
	//"FilterCore":     MakeCoreFilter,
}

func MakeFilters(opts map[string]string) ([]Filter, error) {
	filters := make([]Filter, 0, len(FilterFactory))
	//filters := []Filter{}

	for filterName, makeFunc := range FilterFactory {
		value := opts[filterName]
		if value != "" {
			filter, err := makeFunc(value)
			if err != nil {
				return nil, err
			}
			filters = append(filters, filter)
		}

	}
	return filters, nil

}
