# NameSearch

## Usage
``` go
package main

import (
	"fmt"

	"github.com/liwnn/gopkg/namesearch"
)

func main() {
	ns := namesearch.New()
	ns.Add(1, "abc")
	ns.Add(2, "bcd")
	ns.Add(3, "dd")

	result := ns.Search("bc")
	fmt.Println("Search bc: ", result)

	result = ns.Search("dd")
	fmt.Println("Search dd: ", result)
}
```
