package main

import (
	"fmt"
	"github.com/EilenC/ecommon/bitwise"
	"os"
	"strings"
)

func main() {
	seed := "eilenc"
	prePath, _ := os.Getwd()
	prePath = strings.Join([]string{prePath, "bitwise", "example"}, string(os.PathSeparator))
	fileName := `test.txt`
	encName, err := bitwise.EncryptFile(fileName, seed, prePath)
	if err != nil {
		panic(err)
	}
	decName, err := bitwise.DecryptFile(strings.Join([]string{prePath, encName}, string(os.PathSeparator)), seed, "dec_")
	if err != nil {
		panic(err)
	}
	fmt.Println("enc file : ", encName)
	fmt.Println("dec file : ", decName)
}
