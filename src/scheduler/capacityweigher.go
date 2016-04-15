package scheduler

import (
	"strconv"

	"meta/proto"
)

type CapacityWeigher struct {
	weight float64
}

func (weigher *CapacityWeigher) Weigher(devices []*metaproto.Device) []float64 {
	capacityScore := make([]float64, len(devices))
	allTotal := make([]float64, len(devices))
	for index, v := range devices {
		vTotal, _ := strconv.ParseInt(string(v.Total), 10, 64)
		allTotal[index] = float64(vTotal)
	}
	sumTotal := SumofSliceFloat64(allTotal)

	for index, v := range allTotal {
		capacityScore[index] = sumTotal / v
	}
	capacityScore = Normalization(capacityScore)
	capacityCost := SliceMultiplyFloat64(capacityScore, weigher.weight)

	return capacityCost
}

func MakeCapacityWeigher(value string) (Weigher, error) {
	weight, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}
	var weigherPtr Weigher = &CapacityWeigher{weight}
	return weigherPtr, nil
}
