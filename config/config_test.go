package config

import (
	"io/ioutil"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestConfig(t *testing.T) {

	yamlFile, _ := ioutil.ReadFile("./config.yaml")

	var config Config
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

	if !config.Capability.Video.Rtx {
		t.Error("parse bool error")
	}

}
