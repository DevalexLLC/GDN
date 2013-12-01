
package gdncache

import (
	"fmt"
	"math/rand"
	//"io"
	"crypto/md5"
	"encoding/hex"
)


type Image struct {
	key int
	value string
	hash string
}

type Cache struct {
	cap int
	data map[int]*Image
	keys []int
}

func NewCache (cap int) (*Cache) {
	return &Cache{cap, make(map[int]*Image),
	make([]int, cap)}
}

func (c *Cache) Get(key int) (*Image) {
	return c.data[key]
}

func (c *Cache) Put(key int, value string) {
	h := md5.New()
	h.Write([]byte(value))
	//hash := h.Sum(nil)
	hash := hex.EncodeToString(h.Sum(nil))
	slot := len(c.data)
	if len(c.data) == c.cap {
		// slot = oldest unused key
		// delete(c.data, c.keys[slot])
		slot = rand.Intn(c.cap)
		delete(c.data, c.keys[slot])
		fmt.Printf("Evicted %v \n", c.keys[slot])
	}

	c.keys[slot] = key
	c.data[key] = &Image{key, value, hash}
}
