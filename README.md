# RTClive
A Golang WebRTC/RTMP Low Latency Broadcast Server


# Features

- WebRTC Play
- RTMP Push
- RTMP To WebRTC(audio trancode using ffmpeg)
- WebRTC Server Relay
- Cluster Support 


# Usage

## Run

You should install `ffmpeg`  `media-server-go-native`  and  `media-server-go`  first


[media-server-go](https://github.com/notedit/media-server-go#install)
[media-media-go-native](https://github.com/notedit/media-server-go-native)


```
git clone https://github.com/notedit/rtclive.git

go run server.go  -c config.yaml

```


## Cluster


TBD 









