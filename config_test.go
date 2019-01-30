package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"testing"
	)

func TestConfig(t *testing.T) {

	yamlFile,_ := ioutil.ReadFile("./config.yaml")

	var config ConfigStruct
	err := yaml.Unmarshal(yamlFile, &config)

	if err != nil {
		t.Error("parse yaml error")
	}

	if config.Server.Port != 5000 {
		t.Error("parse int error")
	}

	if len(config.Capability.Video.Rtcpfbcs) != 4 {
		t.Error("parse flow error")
	}

	if !config.Capability.Video.Rtx  {
		t.Error("parse bool error")
	}

}