package metadata

import (
	"meta/proto"
	"path/filepath"
	"store"

	"github.com/golang/protobuf/proto"
)

func getAndDecodeContainer(containerid string) (*metaproto.Container, error) {
	if len(containerid) != CONTAINER_ID_LENGTH {
		return nil, NewError(EcodeParameterError, "container id not valid.")
	}

	containerkey := GenerateContainerKey(containerid)

	driver := store.GetDriver()

	opts := map[string]string{
		"sorted": "false",
		"qurum":  "false",
	}
	data, err := driver.Get(containerkey, opts)
	if err != nil {
		if ValidKeyNotFoundError(err) == true {
			return nil, nil
		}
		return nil, NewError(EcodeBackendError, err.Error())
	}

	ct := &metaproto.Container{}

	err = proto.Unmarshal([]byte(data), ct)
	if err != nil {
		return nil, NewError(EcodeRequestDecodeError, err.Error())
	}

	return ct, err
}

func listContainers() ([]string, error) {
	driver := store.GetDriver()

	containerkey := CONTAINERROOT
	opts := map[string]string{
		"recursive": "false",
		"sorted":    "false",
		"quorum":    "false",
	}
	containers, err := driver.List(containerkey, opts)
	if err != nil {
		if ValidKeyNotFoundError(err) == true {
			return nil, nil
		}
		return nil, NewError(EcodeBackendError, err.Error())
	}

	return containers, nil
}

func setAndEncodeContainer(ct *metaproto.Container) error {
	if ct == nil {
		return NewError(EcodeParameterError, "Container struct not valid.")
	}

	data, err := proto.Marshal(ct)
	if err != nil {
		return NewError(EcodeRequestEncodeError, err.Error())
	}

	driver := store.GetDriver()

	containerkey := GenerateContainerKey(string(ct.Id))
	opts := map[string]string{
		"ttl":       "0",
		"prevValue": "",
		"prevIndex": "0",
	}
	err = driver.Set(containerkey, string(data), opts)
	if err != nil {
		return NewError(EcodeBackendError, err.Error())
	}

	return nil
}

func AddContainer(ct *metaproto.Container) error {
	if ct == nil {
		return NewError(EcodeParameterError, "Container struct not valid.")
	}

	err := setAndEncodeContainer(ct)
	if err != nil {
		return err
	}

	return nil
}

func AddContainerWithVolumes(ct *metaproto.Container, vols []*metaproto.Volume) error {
	if ct == nil {
		return NewError(EcodeParameterError, "Container struct not valid.")
	}
	if len(vols) == 0 {
		return NewError(EcodeParameterError, "No Volume to Add.")
	}

	err := setAndEncodeContainer(ct)
	if err != nil {
		return err
	}

	return nil
}

func GetContainer(containerid string) (*metaproto.Container, error) {
	if len(containerid) != CONTAINER_ID_LENGTH {
		return nil, NewError(EcodeParameterError, "container id not valid.")
	}

	ct, err := getAndDecodeContainer(containerid)
	if err != nil {
		return nil, err
	}

	if string(ct.Id) != containerid {
		log.Errorf("Container key and id not match.")
		//TODO : modify container.id
	}

	return ct, nil
}

func ListContainers() ([]string, error) {
	return listContainers()
}

func ListContainersName() ([]string, error) {
	containers, err := listContainers()
	if err != nil {
		return nil, err
	}

	names := []string{}
	var name string
	for i := 0; i < len(containers); i++ {
		name = filepath.Base(containers[i])
		if 0 != len(name) {
			names = append(names, name)
		}
	}

	return names, nil
}

func DelContainer(containerid string) error {
	if len(containerid) != CONTAINER_ID_LENGTH {
		return NewError(EcodeParameterError, "container id not valid.")
	}

	containerkey := GenerateContainerKey(containerid)

	/*
		ct, err := getAndDecodeContainer(containerid)
		if err != nil {
			log.Errorf("[AddContainer] DelContainer id: %s error: %s", containerid, err.Error())
			return err
		}

		var cvl *metaproto.Container_AttachVolume


			for i := 0; i < len(ct.Volumes); i++ {
				cvl = ct.Volumes[i]

				err = DelVolumeContainer(string(cvl.Volumeid), containerid)

				//TODO record delete fail operation

			}
	*/

	driver := store.GetDriver()

	opts := map[string]string{
		"recursive": "false",
		"dir":       "false",
		"prevValue": "",
		"prevIndex": "0",
	}
	err := driver.Remove(containerkey, opts)
	if err != nil {
		if ValidKeyNotFoundError(err) == true {
			return nil
		}
		return NewError(EcodeBackendError, err.Error())
	}
	return nil
}

func AddContainerVolume(containerid string, volumeid []byte, mode []byte) error {
	if len(containerid) != CONTAINER_ID_LENGTH {
		return NewError(EcodeParameterError, "container id not valid.")
	}
	//TODO check volume exist

	ct, err := GetContainer(containerid)
	if err != nil {
		return err
	}

	var cvl *metaproto.Container_AttachVolume
	for i := 0; i < len(ct.Volumes); i++ {
		cvl = ct.Volumes[i]
		if string(cvl.Volumeid) == string(volumeid) {
			return nil
		}
	}

	avl := &metaproto.Container_AttachVolume{Volumeid: volumeid, Mode: []byte(RWVolume)}

	ct.Volumes = append(ct.Volumes, avl)
	return setAndEncodeContainer(ct)
}

func DelContainerVolume(containerid string, volumeid []byte) error {
	if len(containerid) != CONTAINER_ID_LENGTH {
		return NewError(EcodeParameterError, "container id not valid.")
	}

	ct, err := GetContainer(containerid)
	if err != nil {
		return err
	}

	vols := []*metaproto.Container_AttachVolume{}
	var cvl *metaproto.Container_AttachVolume
	for i := 0; i < len(ct.Volumes); i++ {
		cvl = ct.Volumes[i]
		if string(cvl.Volumeid) == string(volumeid) {
			continue
		}
		vols = append(vols, cvl)
	}

	ct.Volumes = vols

	return setAndEncodeContainer(ct)
}
