package main

import (
	b64 "encoding/base64"
	"fmt"
)

func main(){

	str := "eW91ciB0ZXh0"
	strBytes,_ := b64.StdEncoding.DecodeString(str)
	fmt.Println(strBytes)
}