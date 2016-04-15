package main

import (
	"fmt"
	"store"
	"store/etcd"
)

func test_set() {
	opts := map[string]string{}
	err := store.Backend.Set("kkkkk", "vvvvvvvvvvvvv", opts)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("Set Successfully.")
}

func test_get() {

	opts := map[string]string{}
	resp, err := store.Backend.Get("kkkkk", opts)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(resp)
}

func main() {
	etcd.NewStore()
	test_set()
	test_get()
}
