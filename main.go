package main

import (
	"appinfo_generator/code"
	"fmt"
)

func main() {

	result := code.GeneratAppInfo("com.prime.story.android", 270, 100038, 1001243)

	fmt.Println(result)
}
