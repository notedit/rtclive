package streamer

import (
	gstreamer "github.com/notedit/gstreamer-go"
	mediaserver "github.com/notedit/media-server-go"
)

type WebRTCStreamer struct {
	rtmpUrl string

	audioTrack *mediaserver.IncomingStreamTrack
	videoTrack *mediaserver.IncomingStreamTrack

	audioPipeline *gstreamer.Pipeline
	videoPipeline *gstreamer.Pipeline
	audiosrc      *gstreamer.Element
	videosrc      *gstreamer.Element
}

func NewWebRTCStreamer(rtmpUrl string, audioTrack *mediaserver.IncomingStreamTrack, videoTrack *mediaserver.IncomingStreamTrack) *WebRTCStreamer {

	streamer := &WebRTCStreamer{}
	streamer.rtmpUrl = rtmpUrl

	return nil
}
