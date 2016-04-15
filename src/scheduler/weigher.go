package scheduler

import (
	"meta/proto"
)

type Weigher interface {
	Weigher([]*metaproto.Device) []float64
}

type makeWeigherFunc func(string) (Weigher, error)

var WeigherFactory = map[string]makeWeigherFunc{
	WeigherCapacity: MakeCapacityWeigher,
	//"WeigherCore":     MakeCoreWeigher,
}

func MakeWeighers(opts map[string]string) ([]Weigher, error) {
	weighers := make([]Weigher, 0, len(WeigherFactory))
	//weighers := []Weigher{}

	for weigherName, makeFunc := range WeigherFactory {
		value := opts[weigherName]
		if value != "" {
			weigher, err := makeFunc(value)
			if err != nil {
				return nil, err
			}
			weighers = append(weighers, weigher)
		}

	}
	return weighers, nil

}
