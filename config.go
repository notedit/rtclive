package main



type ConfigStruct struct {
	Server  struct{
		Port  int `yaml:"port"`
	}   `yaml:"server"`

	Cluster struct{
		Origins []string `yaml:"origins,flow"`
	}  `yaml:"cluster"`

	Capability struct{
		Audio  struct{
			Codecs []string    `yaml:"origins,flow"`
			Extensions []string  `yaml:"extensions,flow"`
		}
		Video  struct{
			Codecs []string    `yaml:"origins,flow"`
			Rtx bool            `yaml:"rtx"`
			Extensions []string  `yaml:"extensions,flow"`
			Rtcpfbcs []struct{
				ID  string    `yaml:"id"`
				Params []string `yaml:"params,flow"`
			}  `yaml:"rtcpfbc,flow"`
		}
	}

}
