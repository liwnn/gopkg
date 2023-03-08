# BloomFilter

This is an implementation of BloomFilter written in Go. 

# Usage

``` go
package main

import (
	"fmt"

	"github.com/liwnn/gopkg/bloomfilter"
)

func main() {
	bf := bloomfilter.New(1000, 0.01)
	n1 := []byte("Hurst")
	n2 := []byte("Peek")
	n3 := []byte("Beaty")
	bf.Add(n1)
	bf.Add(n3)
	fmt.Println(bf.MayContain(n1))
	fmt.Println(bf.MayContain(n2))
	fmt.Println(bf.MayContain(n3))
}
```
output
``` go
true
false
true
```
