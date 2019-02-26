# RTClive
A Golang WebRTC/RTMP Broadcast Server



# Features

- WebRTC Push
- WebRTC Play
- RTMP Push
- RTMP To WebRTC(audio trancode using gstreamer)
- WebRTC Server Relay
- Cluster Support 
- Client SDK 


# Todo

- [ ] WebRTC To RTMP
- [ ] HTTP-FLV Push
- [ ] Event Callback
- [ ] Better Cluster Support


# Usage




## Run

```
git clone https://github.com/notedit/RTCLive.git

go run server.go  -c config.yaml

```



## WebRTC Push


See [demo](https://github.com/notedit/RTCLive-js/blob/master/demo/pusher.html)


## RTMP Push

```
ffmpeg -re -i test.mp4  -vcodec copy -acodec copy -f flv rtmp://localhost/live/streamId
```

## WebRTC Play

See [demo](https://github.com/notedit/RTCLive-js/blob/master/demo/player.html)

need change the startPlay's streamId 









