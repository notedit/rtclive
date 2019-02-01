module github.com/notedit/RTCLive

require (
	github.com/gin-contrib/cors v0.0.0-20190101123304-5e7acb10687f
	github.com/gin-contrib/sse v0.0.0-20190125020943-a7658810eb74 // indirect
	github.com/gin-gonic/gin v1.3.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/imroc/req v0.2.3
	github.com/notedit/media-server-go v0.1.5
	gopkg.in/olahol/melody.v1 v1.0.0-20170518105555-d52139073376
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/notedit/media-server-go v0.1.5 => ../../media-server-go
