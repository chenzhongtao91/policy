package main

import (
	"fmt"
	"meta"
	"store/etcd"
)

/*
func test_add_device(devid string, ip string, port int, total int, free int, status int) {
	err := metadata.AddDevice(devid, ip, port, total, free, status)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Add Device Successfully.")
}

func test_get_device(devid string) {
	dv, err := metadata.GetDevice(devid)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(dv)
}

func test_del_device(devid string) {
	err := metadata.DelDevice(devid)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Del Device Successfully.")
}

func test_update_device(devid string, ip string, port int, total int, free int, status int) {
	err := metadata.UpdateDevice(devid, ip, port, total, free, status)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Update Device Successfully.")
}

func test_update_device_net(devid string, ip string, port int) {
	err := metadata.UpdateDeviceNet(devid, ip, port)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Update Device Net Successfully.")
}

func test_update_device_capacity(devid string, total int, free int) {
	err := metadata.UpdateDeviceCapacity(devid, total, free)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Update Device Capacity Successfully.")
}

func test_update_device_status(devid string, status int) {
	err := metadata.UpdateDeviceStatus(devid, status)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Update Device Status Successfully.")
}

func test_list_devices() {
	devices, err := metadata.ListDevices()
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
	fmt.Println(devices)
}

func test_list_device_name() {
	devices, err := metadata.ListDevicesName()
	if err != nil {
		fmt.Println("%s\n", err.Error())
		return
	}
	fmt.Println(devices)
}

func test_use_device(devid string) {
	err := metadata.UseDevice(devid)
	if err != nil {
		fmt.Println("[test_use_device]%s\n", err.Error())
		return
	}

	fmt.Println("Use device id")
}

func test_free_device(devid string) {
	err := metadata.FreeDevice(devid)
	if err != nil {
		fmt.Println("[test_free_device]%s\n", err.Error())
		return
	}

	fmt.Println("Free device id")
}
*/
func test_get_free_device(backend string) {
	devs, err := metadata.GetFreeDevices(backend)
	if err != nil {
		fmt.Println("[test_free_device]%s\n", err.Error())
		return
	}
	fmt.Println(devs)
}

func main() {
	etcd.NewStore()
	/*
		devid := "dev123456789022222222222222222222"
		ip := "192.168.200.200"
		port := int(9510)
		total := int(2048)
		free := int(1530)
		status := int(1)

			test_add_device(devid, ip, port, total, free, status)
			//test_get_device(devid)

					ip = "192.168.222.222"
					port = int(2222)
					total = int(2222)
					free = int(22222)
					//status = int(metadata.DEVICE_INUSE)
					//test_update_device(devid, ip, port, total, free, status)
					fmt.Println("\n########## test_update_device_net #############")
					test_update_device_net(devid, ip, port)
					test_get_device(devid)
					fmt.Println("########## test_update_device_net #############\n")
					fmt.Println("\n########## test_update_device_capacity #############")
					test_update_device_capacity(devid, total, free)
					test_get_device(devid)
					fmt.Println("########## test_update_device_capacity #############\n")
					fmt.Println("\n########## test_update_device_status #############")
					test_update_device_status(devid, status)
					test_get_device(devid)
					fmt.Println("########## test_update_device_status #############\n")

					test_list_devices()
					test_list_device_name()


				test_use_device(devid)
				test_get_device(devid)

				test_free_device(devid)
				test_get_device(devid)
	*/
	test_get_free_device("ceph")

	//test_del_device(devid)
	//test_get_device(devid)

}
