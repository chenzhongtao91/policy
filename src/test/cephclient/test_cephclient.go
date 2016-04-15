package main

import (
	"fmt"

	"cephclient"
)

func main() {
	client, err := cephclient.NewCephClient("192.168.15.11", 6789)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer client.Destroy()
	pools, err := client.ListPools()
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(pools)
	}

	err = client.CreatePool("pool-test")
	if err != nil {
		fmt.Println(err)
	}

	err = client.CreateImage("pool-test", "image1", 1024, 22)
	if err != nil {
		fmt.Println(err)
	}
	err = client.CreateImage("pool-test", "image2", 1024, 22)
	if err != nil {
		fmt.Println(err)
	}
	imageinfo, err := client.ImageStat("pool-test", "image2")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(imageinfo)
	imageName, err := client.ListImageNames("pool-test")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(imageName)

	err = client.CreateSnapshot("pool-test", "image1", "snap1")
	if err != nil {
		fmt.Println(err)
	}

	err = client.CreateSnapshot("pool-test", "image1", "snap2")
	if err != nil {
		fmt.Println(err)
	}

	err = client.Rollback("pool-test", "image1", "snap2")
	if err != nil {
		fmt.Println(err)
	}
	snapNames, err := client.GetSnapshotNames("pool-test", "image1")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(snapNames)

	err = client.Backup("pool-test", "image1", "snap2", "image-backup")
	if err != nil {
		fmt.Println(err)
	}
	imageinfo, err = client.ImageStat("pool-test", "image-backup")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("image-backup info:", imageinfo)

	err = client.RemoveSnapshot("pool-test", "image1", "snap1")
	if err != nil {
		fmt.Println(err)
	}

	err = client.RemoveSnapshot("pool-test", "image1", "snap2")
	if err != nil {
		fmt.Println(err)
	}

	err = client.DeleteImage("pool-test", "image-backup")
	if err != nil {
		fmt.Println(err)
	}

	err = client.DeleteImage("pool-test", "image1")
	if err != nil {
		fmt.Println(err)
	}
	err = client.DeleteImage("pool-test", "image2")
	if err != nil {
		fmt.Println(err)
	}
	err = client.DeletePool("pool-test")
	if err != nil {
		fmt.Println(err)
	}
}
