package main

import (
	"fmt"
	"github.com/shaunhess/gdn/gdncache"
	"strconv"
)

func main() {
	c := gdncache.NewCache(10000)
	i := 0
	for i < 10 {
		c.Put(i, "Image "+strconv.Itoa(i))
		c.Get(i)
		fmt.Printf("Put in cache: %v \n", i)
		fmt.Printf("Testing cache: %v \n", c.Get(i))
		i += 1
	}
}
