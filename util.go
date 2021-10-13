package main

import (
	"io/ioutil"
	"os"

	"github.com/loynoir/ExpandUser.go"
)

func WriteNetRc(netRcContent string) {
	netRcPath, err := ExpandUser.ExpandUser("~/.netrc")
	PanicOnErr(err)
	err = ioutil.WriteFile(netRcPath, []byte(netRcContent), os.ModeAppend)
	PanicOnErr(err)
	Logger.Info(".netrc has been written")
}

func PanicOnErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}
