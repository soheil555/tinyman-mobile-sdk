package main

import (
	"encoding/json"
	"fmt"
)

type A struct {
	Name string
}

func main() {

	a := map[string]map[A]string{
		"test": {
			A{Name: "soheil"}: "test2",
		},
	}

	b, err := json.Marshal(a)

	if err != nil {
		fmt.Println(err)
		return
	}

	var c map[string]map[string]string

	json.Unmarshal(b, &c)

	fmt.Println(c)

}
