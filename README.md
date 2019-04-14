# RTClive
A Golang WebRTC/RTMP Low Latency Broadcast Server


# Features

- WebRTC Push
- WebRTC Play
- RTMP Push
- RTMP To WebRTC(audio trancode using gstreamer)
- WebRTC Server Relay
- Cluster Support 
- Client SDK 


# Usage




## Run


You should install `media-server-go-native`  and `media-server-go`  first


[media-server-go](https://github.com/notedit/media-server-go#install)
[media-media-go-native](https://github.com/notedit/media-server-go-native)

```
git clone https://github.com/notedit/rtclive.git

go run server.go  -c config.yaml

```


## WebRTC Push


See [demo](https://github.com/notedit/RTCLive-js/blob/master/demo/pusher.html)


```
const videoElement = document.getElementById('video_container');

const pusherConfig = new RTCPusherConfig();
const pusher = new RTCPusher(pusherConfig);
await pusher.setupLocalMedia();
pusher.play(videoElement);

const pushUrl =  'ws://127.0.0.1:5000/ws';
await pusher.startPush('test_streamID', pushUrl);
```


## RTMP Push

```
ffmpeg -re -i test.mp4  -vcodec copy -acodec copy -f flv rtmp://localhost/live/streamId
```

## WebRTC Play

See [demo](https://github.com/notedit/RTCLive-js/blob/master/demo/player.html)

```

const videoElement = document.getElementById('video_container');
const playerConfig = new RTCPlayerConfig();
const player = new RTCPlayer(playerConfig);

const playUrl =  'ws://127.0.0.1:5000/ws';
await player.startPlay('test_streamID',playUrl);

player.play(videoElement);
console.log('start to play')
```


## Cluster

rtclive support server relay, when rtclive server can not find one stream, it will find stream from origin servers.

you can config multi origin servers.


```
cluster:
    origins:
        - 127.0.0.1:5001
        - 127.0.0.1:5003

```









