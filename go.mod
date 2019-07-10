module github.com/notedit/rtclive

require (
	github.com/akamensky/argparse v0.0.0-20190115094700-b33e05fb8d69
	github.com/gin-contrib/cors v0.0.0-20190101123304-5e7acb10687f
	github.com/gin-gonic/gin v1.3.0
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/imroc/req v0.2.3
	github.com/notedit/media-server-go v0.1.12
	github.com/notedit/rtmp-lib v0.0.2
	github.com/notedit/sdp v0.0.1
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/notedit/media-server-go v0.1.12 => ../../media-server-go
