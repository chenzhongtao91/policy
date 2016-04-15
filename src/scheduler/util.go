package scheduler

import (
	"fmt"
	"math"
	"strconv"

	"meta/proto"
)

/*
type ByTotal []*metaproto.Device

func (a ByTotal) Len() int      { return len(a) }
func (a ByTotal) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByTotal) Less(i, j int) bool {
	iTotal, _ := strconv.ParseInt(string(a[i].Total), 10, 64)
	jTotal, _ := strconv.ParseInt(string(a[j].Total), 10, 64)
	return iTotal < jTotal
}
*/

type byfunc func(p, q *metaproto.Device) bool

type DeviceWrapper struct {
	device []*metaproto.Device
	by     byfunc
}

type DeviceCostWrapper struct {
	device []*metaproto.Device
	cost   []float64
}

func (dw DeviceWrapper) Len() int {
	return len(dw.device)
}
func (dw DeviceWrapper) Swap(i, j int) {
	dw.device[i], dw.device[j] = dw.device[j], dw.device[i]
}
func (dw DeviceWrapper) Less(i, j int) bool {
	return dw.by(dw.device[i], dw.device[j])
}

func ByTotal(p, q *metaproto.Device) bool {
	iTotal, _ := strconv.ParseInt(string(p.Total), 10, 64)
	jTotal, _ := strconv.ParseInt(string(q.Total), 10, 64)
	return iTotal < jTotal
}

func ReverseByTotal(p, q *metaproto.Device) bool {
	iTotal, _ := strconv.ParseInt(string(p.Total), 10, 64)
	jTotal, _ := strconv.ParseInt(string(q.Total), 10, 64)
	return iTotal > jTotal
}

func (dcw DeviceCostWrapper) Len() int {
	return len(dcw.device)
}
func (dcw DeviceCostWrapper) Swap(i, j int) {
	dcw.device[i], dcw.device[j] = dcw.device[j], dcw.device[i]
	dcw.cost[i], dcw.cost[j] = dcw.cost[j], dcw.cost[i]
}

// 逆序
func (dcw DeviceCostWrapper) Less(i, j int) bool {
	return dcw.cost[i] > dcw.cost[j]
}

func SumofSliceFloat64(data []float64) float64 {
	var sum float64 = 0
	for _, value := range data {
		sum = sum + value
	}
	return sum
}

func MaxofSliceFloat64(data []float64) (float64, int) {
	var max float64 = math.SmallestNonzeroFloat64
	var index int = -1
	for i, value := range data {
		if value > max {
			max = value
			index = i
		}
	}
	return max, index
}

func MinofSliceFloat64(data []float64) (float64, int) {
	var min float64 = math.MaxFloat64
	var index int = -1
	for i, value := range data {
		if value < min {
			min = value
			index = i
		}
	}
	return min, index
}

// 归一化
func Normalization(data []float64) []float64 {
	max, _ := MaxofSliceFloat64(data)
	min, _ := MinofSliceFloat64(data)
	if math.Dim(max, min) < 1e-7 {
		tmpSlice := make([]float64, len(data))
		for i := 0; i < len(tmpSlice); i++ {
			tmpSlice[i] = 1
		}
		return tmpSlice
	} else {
		return SliceMultiplyFloat64(SliceMinusFloat64(data, min), 1/(max-min))
	}
}

func SliceMultiplyFloat64(data []float64, num float64) []float64 {
	newdata := make([]float64, len(data))
	for index, value := range data {
		newdata[index] = value * num
		//newdata = append(newdata, value*num)
	}
	return newdata
}

func SliceMinusFloat64(data []float64, num float64) []float64 {
	newdata := make([]float64, len(data))
	for index, value := range data {
		newdata[index] = value - num
		//newdata = append(newdata, value*num)
	}
	return newdata
}

func SliceAddSliceFloat64(data1 []float64, data2 []float64) ([]float64, error) {
	if len(data1) != len(data2) {
		return nil, fmt.Errorf("Type does not support addition.")
	}
	sumData := make([]float64, len(data1))
	for i := 0; i < len(data1); i++ {
		sumData[i] = data1[i] + data2[i]
	}
	return sumData, nil
}
