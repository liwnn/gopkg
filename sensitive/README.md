# Sensitive

基于AC自动机的屏蔽字检测

# Usage
``` go
package main

import (
	"fmt"

	"github.com/liwnn/gopkg/sensitive"
)

func main() {
	s := sensitive.New() // or sensitive.NewDoubleArray()
	words := []string{"she", "hers", "his"}
	for _, word := range words {
		s.Add(word)
	}
	s.Build()

	if s.Contains("shis") {
		fmt.Println("Contains Success")
	}
	if newWord := s.Replace("shis", '*'); newWord == "s***" {
		fmt.Println("Replace Success")
	}
}
```

output:
```
Contains Success
Replace Success
```
