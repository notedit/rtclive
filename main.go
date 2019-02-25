package main

import (
	"fmt"
	"os"

	"github.com/akamensky/argparse"
	"github.com/notedit/RTCLive/config"
	"github.com/notedit/RTCLive/server"
)

func main() {

	var err error
	var cfg *config.Config
	parser := argparse.NewParser("RTCLive", "RTCLive: WebRTC/RTMP based live streaming server")
	configfile := parser.String("c", "config", &argparse.Options{Required: true, Help: "configpath is required"})

	err = parser.Parse(os.Args)

	if err != nil {
		fmt.Println(err)
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
