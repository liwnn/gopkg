# BitSet
A fast bit array implement in go. Compared with other implementations, It has O(1) time complexity  
for counting the number of 1.  

[![Go Report Card](https://goreportcard.com/badge/github.com/liwnn/gopkg/bitset)](https://goreportcard.com/report/github.com/liwnn/gopkg/bitset)
[![Go Reference](https://pkg.go.dev/badge/github.com/liwnn/gopkg/bitset.svg)](https://pkg.go.dev/github.com/liwnn/gopkg/bitset)

## Usage
``` go
package main

import (
	"fmt"

	"github.com/liwnn/gopkg/bitset"
)

func main() {
	b := bitset.NewSize(8)
	b.Set(1)
	b.Set(100)
	if b.Get(1) {
		fmt.Println("1 is set!")
	}
	if b.Get(100) {
		fmt.Println("100 is set!")
	}

	fmt.Println("Cardinality", b.Cardinality())
	fmt.Println("Length", b.Length())
	fmt.Println("Size", b.Size())

	fmt.Printf("NextSetBit")
	for i, ok := b.NextSetBit(0); ok; i, ok = b.NextSetBit(i + 1) {
		fmt.Printf(" %v", i)
	}
	fmt.Println()

	fmt.Printf("ForeachSetBit")
	b.ForeachSetBit(0, func(j uint) bool {
		fmt.Printf(" %v", j)
		return false
	})
	fmt.Println()

	b.Clear(1)
	if !b.Get(1) {
		fmt.Println("1 is clear!")
	}
}
```
Result:
```
1 is set!
100 is set!     
Cardinality 2   
Length 101      
Size 128        
NextSetBit 1 100
ForeachSetBit 1 100
1 is clear!  
```
