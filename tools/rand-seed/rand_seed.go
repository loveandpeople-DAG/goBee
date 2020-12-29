package main

import (
	"crypto/rand"
	"fmt"

	"github.com/loveandpeople-DAG/goClient/consts"
	"github.com/loveandpeople-DAG/goClient/kerl"
)

func main() {
	b := make([]byte, consts.HashBytesSize)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	// convert to trytes and set the last trit to zero
	seed, err := kerl.KerlBytesToTrytes(b)
	if err != nil {
		fmt.Printf("xxxxxxxxxxxxxx")
		panic(err)
	}

	fmt.Println(seed)
}
