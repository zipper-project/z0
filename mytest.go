package main

import (
	"fmt"

	"github.com/zipper-project/z0/common"
)

type T struct {
	v common.Hash
}

func main() {
	t := T{}
	fmt.Println(t.v[:])

	var a []byte
	if a == nil{
		fmt.Println("hello")
	}
}
