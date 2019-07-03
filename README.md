# RTClive
A Golang WebRTC/RTMP Low Latency Broadcast Server


# Features

- WebRTC Play
- RTMP Push
- RTMP To WebRTC(audio trancode using gstreamer)
- WebRTC Server Relay
- Cluster Support 


# Usage




## Run


Install gstreamer 



ubuntu
```
sudo apt-get install libgstreamer1.0-dev libgstreamer-plugins-base1.0-dev libgstreamer-plugins-good1.0-dev libgstreamer-plugins-bad1.0-dev gstreamer1.0-plugins-ugly gstreamer1.0-libav 
```


mac
```
brew install pkg-config
brew install libffi
brew install gstreamer
brew install gst-plugins-base
brew install gst-plugins-good
brew install gst-plugins-bad
brew install gst-plugins-ugly
export PKG_CONFIG_PATH="/usr/local/opt/libffi/lib/pkgconfig"
```




You should install `media-server-go-native`  and `media-server-go`  first


[media-server-go](https://github.com/notedit/media-server-go#install)
[media-media-go-native](https://github.com/notedit/media-server-go-native)



```
git clone https://github.com/notedit/rtclive.git

go run server.go  -c config.yaml

```



## Cluster


TBD 









