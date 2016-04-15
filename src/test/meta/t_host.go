package main

import (
	"fmt"
	"meta"

	"store/etcd"
)

func test_add_host(ip string) {
	devs := [][]byte{[]byte("sdb"), []byte("sdc"), []byte("sdd")}
	err := metadata.AddHost(ip, 1, devs)
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
	fmt.Println("Add Host Successfully.")
}

func test_get_host(ip string) {
	hs, err := metadata.GetHost(ip)
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
	fmt.Println(hs)
}

func test_del_host() {
	err := metadata.DelHost("192.168.100.100")
	if err != nil {
		fmt.Println("%s\n", err.Error())
	}
}

func test_mod_host_ip(oldip string, newip string) {
	err := metadata.ModHostIp(oldip, newip)
	if err != nil {
		fmt.Println("%s\n", err.Error())
	}
}

func test_mod_host_status(ip string, status int) {
	err := metadata.ModHostStatus(ip, status)
	if err != nil {
		fmt.Println("%s\n", err.Error())
	}
}

func test_add_host_devices(ip string, adddevs [][]byte) {
	err := metadata.AddHostDevices(ip, adddevs)
	if err != nil {
		fmt.Println("%s\n", err.Error())
	}
}

func test_del_host_devices(ip string, deldevs [][]byte) {
	err := metadata.DelHostDevices(ip, deldevs)
	if err != nil {
		fmt.Println("%s\n", err.Error())
	}
}

func test_list_host() {
	hosts, err := metadata.ListHosts()
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
	fmt.Println(hosts)
}

func test_list_host_name() {
	hosts, err := metadata.ListHostsName()
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
	fmt.Println(hosts)
}

func main() {
	etcd.NewStore()
	oldip := "192.168.100.100"
	newip := "192.168.200.200"
	test_add_host(oldip)
	test_get_host(oldip)
	//test_del_host()
	test_mod_host_ip(oldip, newip)
	test_get_host(newip)
	test_mod_host_status(newip, 2)
	test_get_host(newip)
	test_add_host_devices(newip, [][]byte{[]byte("sdx"), []byte("sdy"), []byte("sdz")})
	test_get_host(newip)
	test_del_host_devices(newip, [][]byte{[]byte("sdx"), []byte("sdy")})
	test_get_host(newip)

	//newip = oldip
	//oldip = newip

	test_list_host_name()
}
