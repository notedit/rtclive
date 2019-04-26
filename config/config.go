package config

import (
	"errors"
	"io/ioutil"

	"github.com/notedit/sdp"
	"gopkg.in/yaml.v2"
)


type RelayStruct struct {
	Mode string `yaml:"mode"`
	Type string `yaml:"type"`
	Edge string `yaml:"edge"`
	Stream string `yaml:"stream"`
}

// Config lib 
type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`

	Media struct {
		Endpoint string `yaml:"endpoint"`
	} `yaml:"media"`

	Cluster struct {
		Origins []string `yaml:"origins,flow"`
	} `yaml:"cluster"`


	Relays []RelayStruct `yaml:"relay,flow"`

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

	StaticRelays map[string]*RelayStruct
	DynamicRelay *RelayStruct
	AudioCapability *sdp.Capability
	VideoCapability *sdp.Capability
	Capabilities map[string]*sdp.Capability
}

// LoadConfigBytes load config from data buffer
func LoadConfigBytes(data []byte) (*Config, error) {

	var config Config
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	config.StaticRelays = make(map[string]*RelayStruct)

	for _,relay := range config.Relays {
		if relay.Type == "static" {
			if relay.Stream == "" {
				return nil, errors.New("static relay should have stream name")
			}
			config.StaticRelays[relay.Stream] = &relay
		}
		if relay.Type == "dynamic" {
			if config.DynamicRelay != nil {
				return nil, errors.New("dynamic relay should just have one ")
			}
			config.DynamicRelay = &relay
		}
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
		config.AudioCapability = audioCapability
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
		config.VideoCapability = videoCapability
		config.Capabilities["video"] = videoCapability
	}

	return &config, nil
	
}


// LoadConfig from a file 
func LoadConfig(filePath string) (*Config, error) {

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return LoadConfigBytes(data)
}
