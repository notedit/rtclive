module github.com/notedit/RTCLive

require (
	github.com/gin-contrib/cors v0.0.0-20190101123304-5e7acb10687f
	github.com/gin-contrib/sse v0.0.0-20190125020943-a7658810eb74 // indirect
	github.com/gin-gonic/gin v1.3.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/hashicorp/go-multierror v1.0.0 // indirect
	github.com/imroc/req v0.2.3
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/nareix/joy4 v0.0.0-20181022032202-3ddbc8f9d431
	github.com/notedit/gstreamer-go v0.2.0
	github.com/notedit/media-server-go v0.1.5
	github.com/pkg/errors v0.8.1 // indirect
	github.com/sirupsen/logrus v1.3.0 // indirect
	github.com/yutopp/go-amf0 v0.0.0-20180803120851-48851794bb1f // indirect
	github.com/yutopp/go-flv v0.2.0
	github.com/yutopp/go-rtmp v0.0.0-20190128071726-cb2d763d2aac
	gopkg.in/olahol/melody.v1 v1.0.0-20170518105555-d52139073376
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/notedit/media-server-go v0.1.5 => ../../media-server-go
