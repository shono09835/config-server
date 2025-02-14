package main

import (
	"fmt"
	"os"

	"github.com/shono09835/config-server/config"
	"github.com/shono09835/config-server/log"
	"github.com/shono09835/config-server/server"
)

func main() {
	defer log.Logger.HandlePanic("Main")

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <config-file>\n", os.Args[0])
		os.Exit(1)
	}

	config, err := config.ParseConfig(os.Args[1])
	if err != nil {
		panic("Unable to parse configuration file\n" + err.Error())
	}

	server := server.NewConfigServer(config)
	err = server.Start()
	if err != nil {
		panic("Unable to start server\n" + err.Error())
	}
}
