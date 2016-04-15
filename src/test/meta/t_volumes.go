package main

import (
	"fmt"
	"meta"
	"store/etcd"

	"meta/proto"
)

func test_add_volume(volumeid string, wcontainerid []byte, containers [][]byte, status int) {
	driverName := "ceph"
	vl := &metaproto.Volume{
		Id:       []byte(volumeid),
		Status:   metadata.IntegerToBytes(status),
		Writable: wcontainerid,
	}

	cons := []*metaproto.Volume_OwnerContainer{}

	var oc *metaproto.Volume_OwnerContainer
	for i := 0; i < len(containers); i++ {
		oc = &metaproto.Volume_OwnerContainer{Containerid: containers[i], Mode: []byte(metadata.ROVolume)}

		cons = append(cons, oc)
	}

	cons = append(cons, &metaproto.Volume_OwnerContainer{Containerid: wcontainerid, Mode: []byte(metadata.RWVolume)})

	vl.Containers = cons

	err := metadata.AddVolume(vl, driverName)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func test_get_volume(volumeid string) {
	driverName := "ceph"
	vl, err := metadata.GetVolume(volumeid, driverName)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(vl)
}

func test_del_volume(volumeid string) {
	driverName := "ceph"
	err := metadata.DelVolume(volumeid, driverName)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func test_get_writable_container(volumeid string) {
	driverName := "ceph"
	wc, err := metadata.GetVolumeWRContainer(volumeid, driverName)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Writable Container: ", string(wc))
}

func test_get_readonly_containers(volumeid string) {
	driverName := "ceph"
	ro, err := metadata.GetVolumeROContainers(volumeid, driverName)
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
	for i := 0; i < len(ro); i++ {
		fmt.Println("ReadOnly Container: ", string(ro[i]))
	}

}

func test_list_volumes() {
	driverName := "ceph"
	volumes, err := metadata.ListVolumes(driverName)
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
	fmt.Println(volumes)
}

func test_list_volume_name() {
	//driverName := "ceph"
	devices, err := metadata.ListDevicesName()
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
	fmt.Println(devices)
}

func test_set_volume_container(volumeid string, vct *metaproto.Volume_OwnerContainer, force bool) {
	driverName := "ceph"
	err := metadata.SetVolumeContainer(volumeid, vct, driverName, force)
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
}

func test_del_volume_container(volumeid string, containerid string) {
	driverName := "ceph"
	err := metadata.DelVolumeContainer(volumeid, driverName, containerid)
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
}

func main() {
	etcd.NewStore()
	volumeid := "volume1234567890"
	wrcontainerid := []byte("containerwwww111111111111")
	rocontainers := [][]byte{[]byte("containerrrrr111111111111"), []byte("containerrrrr22222222222"), []byte("containerrrrr333333333")}
	status := 1

	test_add_volume(volumeid, wrcontainerid, rocontainers, status)
	test_get_volume(volumeid)

	//test_get_writable_container(volumeid)
	//test_get_readonly_containers(volumeid)

	new_wcontainer := []byte("container2222222222222222222222222222")
	oc := &metaproto.Volume_OwnerContainer{
		Containerid: new_wcontainer,
		Mode:        []byte(metadata.ROVolume),
	}
	test_set_volume_container(volumeid, oc, false)
	test_get_volume(volumeid)

	new_wcontainer = []byte("container33333333333333333333333333333")
	oc = &metaproto.Volume_OwnerContainer{
		Containerid: new_wcontainer,
		Mode:        []byte(metadata.RWVolume),
	}
	test_set_volume_container(volumeid, oc, false)
	test_get_volume(volumeid)

	new_wcontainer = []byte("container444444444444444444444444444444")
	oc = &metaproto.Volume_OwnerContainer{
		Containerid: new_wcontainer,
		Mode:        []byte(metadata.RWVolume),
	}
	test_set_volume_container(volumeid, oc, true)
	test_get_volume(volumeid)

	fmt.Println("\n\n\n")

	new_wcontainer = []byte("container444444444444444444444444444444")
	test_del_volume_container(volumeid, string(new_wcontainer))
	test_get_writable_container(volumeid)
	test_get_readonly_containers(volumeid)
	test_get_volume(volumeid)

	new_wcontainer = []byte("container2222222222222222222222222222")
	test_del_volume_container(volumeid, string(new_wcontainer))
	test_get_writable_container(volumeid)
	test_get_readonly_containers(volumeid)

	new_wcontainer = []byte("container33333333333333333333333333333")
	test_del_volume_container(volumeid, string(new_wcontainer))
	test_get_writable_container(volumeid)
	test_get_readonly_containers(volumeid)

	//new_rocontainers := [][]byte{[]byte("containerrrrr33333333"), []byte("containerrrrr444444444444"), []byte("containerrrrr5555555555555")}

	//test_add_readonly_containers(volumeid, new_rocontainers)
	//fmt.Println("\n\n")

	//del_rocontainers := [][]byte{[]byte("containerrrrr33333333"), []byte("containerrrrr444444444444")}
	//test_del_volume_readonly_container(volumeid, rocontainers)
	//fmt.Println("\n\n")
	//test_get_readonly_containers(volumeid)

	/*
		test_get_writable_container(volumeid)
		test_del_volume_writable_container(volumeid, new_wcontainer)
		test_get_writable_container(volumeid)
		test_del_volume_writable_container(volumeid, wrcontainerid)
		test_get_writable_container(volumeid)
	*/

	//test_list_volumes()
	//test_list_volume_name()

	//test_del_volume(volumeid)
	//test_get_volume(volumeid)
}
