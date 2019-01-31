package main

import (
	"errors"
	"io/ioutil"

	"github.com/notedit/media-server-go/sdp"
	"gopkg.in/yaml.v2"
)

type ConfigStruct struct {
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
}

func LoadConfig(filePath string) (*ConfigStruct, error) {

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config ConfigStruct
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	if len(config.Capability.Audio.Codecs) == 0 && len(config.Capability.Video.Codecs) == 0 {
		return nil, errors.New("capability can not be empty")
	}

	return &config, nil
}

func (c *ConfigStruct) GetCapabilitys() map[string]*sdp.Capability {

	capabilitys := map[string]*sdp.Capability{}

	if c.Capability.Audio.Codecs != nil {
		audioCapability := &sdp.Capability{
			Codecs:     c.Capability.Audio.Codecs,
			Extensions: c.Capability.Audio.Extensions,
		}
		capabilitys["audio"] = audioCapability
	}

	if c.Capability.Video.Codecs != nil {
		rtcpfbs := make([]*sdp.RtcpFeedback, 0)
		for _, rtcpfb := range c.Capability.Video.Rtcpfbcs {
			rtcpfbs = append(rtcpfbs, &sdp.RtcpFeedback{
				ID:     rtcpfb.ID,
				Params: rtcpfb.Params,
			})
		}
		videoCapability := &sdp.Capability{
			Codecs:     c.Capability.Video.Codecs,
			Rtx:        c.Capability.Video.Rtx,
			Extensions: c.Capability.Video.Extensions,
			Rtcpfbs:    rtcpfbs,
		}
		capabilitys["video"] = videoCapability
	}

	return capabilitys
}
