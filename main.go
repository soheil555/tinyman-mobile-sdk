package main

import (
	"fmt"
	"sort"
)

func main() {

	assets := []uint64{2, 1}

	sort.Slice(assets, func(i, j int) bool { return assets[i] < assets[j] })

	fmt.Println(assets[0], assets[1])

}
