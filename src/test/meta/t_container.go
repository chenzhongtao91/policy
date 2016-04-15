package main

import (
	"fmt"
	"meta"
	"store/etcd"

	"meta/proto"
)

func test_add_container(containerid []byte) {
	ct := &metaproto.Container{}

	ct.Id = containerid
	ct.Status = metadata.IntegerToBytes(metadata.CONTAINERONLINE)

	err := metadata.AddContainer(ct)
	if err != nil {
		fmt.Println("%s\n", err.Error())
	}
}

func test_get_container(containerid string) {
	ct, err := metadata.GetContainer(containerid)
	if err != nil {
		fmt.Println("%s\n", err.Error())
	}

	fmt.Println(ct)
}

func test_del_container(containerid string) {
	err := metadata.DelContainer(containerid)
	if err != nil {
		fmt.Println("%s\n", err.Error())
	}
}

func test_list_containers() {
	containers, err := metadata.ListContainers()
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
	fmt.Println(containers)
}

func test_list_container_name() {
	containers, err := metadata.ListContainersName()
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
	fmt.Println(containers)
}

func test_add_container_volume(containerid string, volumeid []byte, mode []byte) {
	err := metadata.AddContainerVolume(string(containerid), volumeid, []byte(metadata.ROVolume))
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
}

func test_del_container_volume(containerid string, volumeid []byte) {
	err := metadata.DelContainerVolume(containerid, volumeid)
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
}

func main() {

	etcd.NewStore()

	containerid := []byte("container22222222222222222222222222")
	//volumes := [][]byte{[]byte("virdisk1"), []byte("virdisk2"), []byte("virdisk3")}
	//status := 1

	test_add_container(containerid)
	test_get_container(string(containerid))
	//test_del_container(string(containerid))
	//test_get_container(string(containerid))
	//test_list_containers()
	//test_list_container_name()

	volume := []byte("virdisk1")
	test_add_container_volume(string(containerid), volume, []byte(metadata.ROVolume))
	test_get_container(string(containerid))

	volume = []byte("virdisk2")
	test_add_container_volume(string(containerid), volume, []byte(metadata.ROVolume))
	test_get_container(string(containerid))

	volume = []byte("virdisk3")
	test_add_container_volume(string(containerid), volume, []byte(metadata.ROVolume))
	test_get_container(string(containerid))

	volume = []byte("virdisk2")
	test_del_container_volume(string(containerid), volume)
	test_get_container(string(containerid))

	volume = []byte("virdisk3")
	test_del_container_volume(string(containerid), volume)
	test_get_container(string(containerid))

}
