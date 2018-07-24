package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
)

func main() {
	fmt.Println("Hi there")
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	str := string(data)
	str = strings.Trim(str, " \n\"'")
	quantity := resource.MustParse(str)

	fmt.Println(quantity.Value())
}
