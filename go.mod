module github.com/notedit/RTCLive

require (
	github.com/gin-contrib/sse v0.0.0-20190125020943-a7658810eb74 // indirect
	github.com/gin-gonic/gin v1.3.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/golang/protobuf v1.2.0 // indirect
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/mattn/go-isatty v0.0.4 // indirect
	github.com/notedit/media-server-go v0.1.4
	github.com/ugorji/go/codec v0.0.0-20181209151446-772ced7fd4c2 // indirect
	gopkg.in/go-playground/validator.v8 v8.18.2 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
)

replace github.com/notedit/media-server-go v0.1.4 => ../../media-server-go
