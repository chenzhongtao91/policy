package daemon

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	//"path"
	//"path/filepath"
	"reflect"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

func LoadConfig(fileName string, v interface{}) error {
	if _, err := os.Stat(fileName); err != nil {
		return err
	}

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	defer file.Close()

	if err = json.NewDecoder(file).Decode(v); err != nil {
		return err
	}
	return nil
}

func SaveConfig(fileName string, v interface{}) error {
	tmpFileName := fileName + ".tmp"

	f, err := os.Create(tmpFileName)
	if err != nil {
		return err
	}

	if err := json.NewEncoder(f).Encode(v); err != nil {
		f.Close()
		return err
	}
	f.Close()

	if _, err = os.Stat(fileName); err == nil {
		if err = os.Remove(fileName); err != nil {
			return err
		}
	}

	if err := os.Rename(tmpFileName, fileName); err != nil {
		return err
	}

	return nil
}

func ConfigExists(fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil
}

func Execute(binary string, args []string) (string, error) {
	var output []byte
	var err error
	cmd := exec.Command(binary, args...)
	done := make(chan struct{})

	go func() {
		output, err = cmd.CombinedOutput()
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(1 * time.Minute):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return "", fmt.Errorf("Timeout executing: %v %v, output %v, error %v", binary, args, string(output), err)
	}

	if err != nil {
		return "", fmt.Errorf("Failed to execute: %v %v, output %v, error %v", binary, args, string(output), err)
	}
	return string(output), nil
}

func RemoveConfig(fileName string) error {
	if _, err := Execute("rm", []string{"-f", fileName}); err != nil {
		return err
	}
	return nil
}

type ObjectOperations interface {
	ConfigFile() (string, error)
}

func getObjectOps(obj interface{}) (ObjectOperations, error) {
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return nil, fmt.Errorf("BUG: Non-pointer was passwd in")
	}
	t := reflect.TypeOf(obj).Elem()
	ops, ok := obj.(ObjectOperations)
	if !ok {
		return nil, fmt.Errorf("BUG: %v doesn't implement necessary methods for accessing object", t)
	}
	return ops, nil
}

func ObjectConfig(obj interface{}) (string, error) {
	ops, err := getObjectOps(obj)
	if err != nil {
		return "", err
	}
	config, err := ops.ConfigFile()
	if err != nil {
		return "", err
	}
	return config, nil
}

func ObjectLoad(obj interface{}) error {
	config, err := ObjectConfig(obj)
	if err != nil {
		return err
	}
	if !ConfigExists(config) {
		return fmt.Errorf("Cannot find object config %v", config)
	}
	if err := LoadConfig(config, obj); err != nil {
		return err
	}
	return nil
}

func ObjectExists(obj interface{}) (bool, error) {
	config, err := ObjectConfig(obj)
	if err != nil {
		return false, err
	}
	return ConfigExists(config), nil
}

func ObjectSave(obj interface{}) error {
	config, err := ObjectConfig(obj)
	if err != nil {
		return err
	}
	return SaveConfig(config, obj)
}

func ObjectDelete(obj interface{}) error {
	config, err := ObjectConfig(obj)
	if err != nil {
		return err
	}
	return RemoveConfig(config)
}

func SliceToMap(slices []string) map[string]string {
	result := map[string]string{}
	for _, v := range slices {
		pair := strings.Split(v, "=")
		if len(pair) != 2 {
			return nil
		}
		result[pair[0]] = pair[1]
	}
	return result
}

func MkdirIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, os.ModeDir|0700); err != nil {
			return err
		}
	}
	return nil
}

func CheckIP(ip string) []byte {
	return net.ParseIP(ip)
}

func LockFile(fileName string) (*os.File, error) {
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func UnlockFile(f *os.File) error {
	defer f.Close()
	if err := unix.Flock(int(f.Fd()), unix.LOCK_UN); err != nil {
		return err
	}
	if _, err := Execute("rm", []string{f.Name()}); err != nil {
		return err
	}
	return nil
}

func decodeRequest(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func EncodeData(v interface{}) (*bytes.Buffer, error) {
	param := bytes.NewBuffer(nil)
	j, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if _, err := param.Write(j); err != nil {
		return nil, err
	}
	return param, nil
}

func ByteArrayToStringArray(b [][]byte) []string {
	var s []string
	for i := 0; i < len(b); i++ {
		s = append(s, string(b[i]))
	}
	return s
}
