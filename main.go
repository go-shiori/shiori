package main

import "fmt"

func main() {
	fmt.Println("Hello world")
	_, err := openDatabase()
	checkError(err)
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
