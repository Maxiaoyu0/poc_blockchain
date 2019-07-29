package main

import (
	"fmt"
	"github.com/fabric-dev-servers-multipeer/sdkInit"
)

const (
	configFile = "config.yaml"
	initialized = false
	//SimpleCC = "simplecc"
)

func main() {


	sdk, err := sdkInit.SetupSDK(configFile, initialized)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	defer sdk.Close()

}