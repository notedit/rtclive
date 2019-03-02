module github.com/notedit/RTCLive

require (
	github.com/akamensky/argparse v0.0.0-20190115094700-b33e05fb8d69
	github.com/gin-contrib/cors v0.0.0-20190101123304-5e7acb10687f
	github.com/gin-contrib/sse v0.0.0-20190125020943-a7658810eb74 // indirect
	github.com/gin-gonic/gin v1.3.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/hajimehoshi/oto v0.3.0
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/imroc/req v0.2.3
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/notedit/go-opus v0.0.0-20190301132009-f6383a937478
	github.com/notedit/go-sdp v0.0.0-20181103104252-cc0b89e031ad
	github.com/notedit/go-sdp-transform v0.0.0-20181119121630-e59ee064108d // indirect
	github.com/notedit/gstreamer-go v0.2.0
	github.com/notedit/media-server-go v0.1.5
	github.com/notedit/rtmp-lib v0.0.1
	github.com/olahol/melody v0.0.0-20180227134253-7bd65910e5ab
	github.com/pkg/errors v0.8.1 // indirect
	github.com/sirupsen/logrus v1.3.0 // indirect
	github.com/winlinvip/go-opus v0.0.0-20180718015314-6724a8a7295a
	gopkg.in/hraban/opus.v2 v2.0.0-20180426093920-0f2e0b4fc6cd
	gopkg.in/olahol/melody.v1 v1.0.0-20170518105555-d52139073376
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/notedit/media-server-go v0.1.5 => ../../media-server-go
