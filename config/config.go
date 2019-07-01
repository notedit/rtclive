package config

import (
	"errors"
	"io/ioutil"

	"github.com/notedit/sdp"
	"gopkg.in/yaml.v2"
)

type serverstruct struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type mediastruct struct {
	Endpoint string `yaml:"endpoint"`
	Minport  int    `yaml:"minport"`
	Maxport  int    `yaml:"maxport"`
}

type relaystruct struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type rtmpstruct struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// Config struct
type Config struct {
	Server     *serverstruct `yaml:"server"`
	Media      *mediastruct  `yaml:"media"`
	Relay      *relaystruct  `yaml:"relay"`
	Rtmp       *rtmpstruct   `yaml:"rtmp"`
	Capability struct {
		Audio struct {
			Codecs     []string `yaml:"codecs,flow"`
			Extensions []string `yaml:"extensions,flow"`
		} `yaml:"audio"`
		Video struct {
			Codecs     []string `yaml:"codecs,flow"`
			Rtx        bool     `yaml:"rtx"`
			Extensions []string `yaml:"extensions,flow"`
			Rtcpfbcs   []struct {
				ID     string   `yaml:"id"`
				Params []string `yaml:"params,flow"`
			} `yaml:"rtcpfbc,flow"`
		} `yaml:"video"`
	}
	Capabilities map[string]*sdp.Capability
}

func LoadConfig(filePath string) (*Config, error) {

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if len(config.Capability.Audio.Codecs) == 0 && len(config.Capability.Video.Codecs) == 0 {
		return nil, errors.New("capability can not be empty")
	}

	config.Capabilities = make(map[string]*sdp.Capability)

	if config.Capability.Audio.Codecs != nil {
		audioCapability := &sdp.Capability{
			Codecs:     config.Capability.Audio.Codecs,
			Extensions: config.Capability.Audio.Extensions,
		}
		config.Capabilities["audio"] = audioCapability
	}

	if config.Capability.Video.Codecs != nil {
		rtcpfbs := make([]*sdp.RtcpFeedback, 0)
		for _, rtcpfb := range config.Capability.Video.Rtcpfbcs {
			rtcpfbs = append(rtcpfbs, &sdp.RtcpFeedback{
				ID:     rtcpfb.ID,
				Params: rtcpfb.Params,
			})
		}
		videoCapability := &sdp.Capability{
			Codecs:     config.Capability.Video.Codecs,
			Rtx:        config.Capability.Video.Rtx,
			Extensions: config.Capability.Video.Extensions,
			Rtcpfbs:    rtcpfbs,
		}
		config.Capabilities["video"] = videoCapability
	}

	return &config, nil
}
