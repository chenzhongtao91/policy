package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var errorStatus = map[int]int{
	EcodeHostNotFound:      http.StatusNotFound,
	EcodeDeviceNotFound:    http.StatusNotFound,
	EcodeContainerNotFound: http.StatusNotFound,
	EcodeVolumeNotFound:    http.StatusNotFound,
	EcodeNotFile:           http.StatusForbidden,
	EcodeDirNotEmpty:       http.StatusForbidden,
	EcodeUnauthorized:      http.StatusUnauthorized,
	EcodeTestFailed:        http.StatusPreconditionFailed,
	EcodeNodeExist:         http.StatusPreconditionFailed,
	EcodeRaftInternal:      http.StatusInternalServerError,
	EcodeLeaderElect:       http.StatusInternalServerError,
}
