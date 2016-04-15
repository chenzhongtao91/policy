package metadata

import (
	"encoding/json"
)

const (
	// Host
	EcodeHostNotFound = 1000
	EcodeHostExist    = 1001

	// Device
	EcodeDeviceNotFound     = 2000
	EcodeDeviceExist        = 2001
	EcodeDeviceToHostsError = 2002
	EcodeDeviceAddError     = 2003
	EcodeDeviceInUse        = 2004

	// Container
	EcodeContainerNotFound = 3000
	EcodeVolumeExist       = 3001
	EcodeVolumeConflict    = 3002

	// Volume
	EcodeVolumeNotFound   = 4000
	EcodeWRContainerExist = 4001
	EcodeVolumeInUse      = 4002
	EcodeVolumeDeviceMiss = 4003

	//Common
	EcodeParameterError     = 5000
	EcodeRequestDecodeError = 5001
	EcodeRequestEncodeError = 5002
	EcodeBackendError       = 5003
	EcodeSchedulerError     = 5004
	EcodeEventTimeExipre    = 5005
	EcodeEventTimeInvalid   = 5006
	EcodeMetaTimeInvalid    = 5007
)

type Error struct {
	Code    int    `json:"errorCode"`
	Message string `json:"message"`
}

func NewError(errorCode int, msg string) *Error {
	return &Error{
		Code:    errorCode,
		Message: msg,
	}
}

// Error is for the error interface
func (e Error) Error() string {
	return e.Message + " (" + string(e.Code) + ")"
}

func (e Error) toJsonString() string {
	b, _ := json.Marshal(e)
	return string(b)
}
