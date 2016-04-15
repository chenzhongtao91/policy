package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
)

type ErrorResponse struct {
	Error string
}

type DeviceIdentify struct {
	IP   string
	Port string
	Dev  string
}

type VolumeResponse struct {
	Result     string
	ID         string
	Status     string
	Capacity   string
	Writable   string
	Containers []string
	Devices    []DeviceIdentify
}

type VolumeListResponse struct {
	Result  string
	Volumes []string
}

type HostResponse struct {
	Result string
	IP     string
	Status string
	Devs   []string
}

type HostListResponse struct {
	Result string
	IPs    []string
}

type DeviceResponse struct {
	Result   string
	ID       string
	IP       string
	Capacity string
	Resource string
}

type DeviceListResponse struct {
	Result  string
	Devices []string
}

//ResponseError would generate a error information in JSON format for output
func ResponseError(format string, a ...interface{}) {
	response := ErrorResponse{Error: fmt.Sprintf(format, a...)}
	j, err := json.MarshalIndent(&response, "", "\t")
	if err != nil {
		panic(fmt.Sprintf("Failed to generate response for error:", err))
	}
	fmt.Println(string(j[:]))
}

//ResponseLogAndError would log the error before call ResponseError()
func ResponseLogAndError(v interface{}) {
	if e, ok := v.(*logrus.Entry); ok {
		e.Error(e.Message)
		oldFormatter := e.Logger.Formatter
		logrus.SetFormatter(&logrus.JSONFormatter{})
		s, err := e.String()
		logrus.SetFormatter(oldFormatter)
		if err != nil {
			ResponseError(err.Error())
			return
		}
		//Cosmetic since " would be escaped
		ResponseError(strings.Replace(s, "\"", "'", -1))
	} else if e, ok := v.(error); ok {
		logrus.Errorf(fmt.Sprint(e))
		ResponseError("Caught FATAL error: %s", v)
	}
}

// ResponseOutput would generate a JSON format byte array of object for output
func ResponseOutput(v interface{}) ([]byte, error) {
	j, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return nil, err
	}
	return j, nil
}
