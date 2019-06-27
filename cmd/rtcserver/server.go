package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/notedit/rtclive/config"
	"github.com/notedit/rtclive/server"
)

func main() {

	var err error
	var cfg *config.Config
	parser := argparse.NewParser("rtclive", "rtclive: webrtc edge live streaming server")
	configfile := parser.String("c", "config", &argparse.Options{Required: false, Help: "config file path", Default: "config.yaml"})

	err = parser.Parse(os.Args)

	if err != nil {
		fmt.Println(parser.Usage(err))
		return
	}

	cfg, err = config.LoadConfig(*configfile)

	if err != nil {
		fmt.Println(err)
		return
	}

	serv := server.New(cfg)

	serv.ListenAndServe()

}
