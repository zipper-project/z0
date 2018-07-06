package main

import (
	"github.com/hashicorp/golang-lru"
	"fmt"
)

func main() {
	lmap, _ := lru.New(100)
	lmap.Add(1, "hello")
	lmap.Add(2, "world")

	for _, k := range lmap.Keys() {
		v, _ := lmap.Get(k)
		fmt.Println(k, v)
	}
}
