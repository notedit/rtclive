package streamer

import (
	gstreamer "github.com/notedit/gstreamer-go"
	mediaserver "github.com/notedit/media-server-go"
)

// flvmux  streamable=true

const rtp2rtmp = "appsrc is-live=true do-timestamp=true name=audiosrc ! decodebin ! audioconvert ! audioresample ! audio/x-raw,rate=441000 ! faac ! mux.  videotestsrc num-buffers=500 ! video/x-raw,framerate=25/1 ! x264enc ! mux.  flvmux name=mux ! filesink location=test.flv"

type WebRTCStreamer struct {
	rtmpUrl string

	audioTrack *mediaserver.IncomingStreamTrack
	videoTrack *mediaserver.IncomingStreamTrack

	pipeline *gstreamer.Pipeline
	audiosrc *gstreamer.Element
	videosrc *gstreamer.Element
}

func NewWebRTCStreamer(rtmpUrl string, audioTrack *mediaserver.IncomingStreamTrack, videoTrack *mediaserver.IncomingStreamTrack) *WebRTCStreamer {

	streamer := &WebRTCStreamer{}
	streamer.rtmpUrl = rtmpUrl

	streamer.audioTrack = audioTrack
	streamer.videoTrack = videoTrack

	return streamer
}

func (self *WebRTCStreamer) setupPipeline() {

	pipeline, err := gstreamer.New(rtp2rtmp)
	if err != nil {
		panic(err)
	}

	self.pipeline = pipeline
	self.audiosrc = pipeline.FindElement("audiosrc")
	self.pipeline.Start()
}

func (self *WebRTCStreamer) PushAudioFrame(data []byte, timestamp uint) {

	if self.pipeline == nil {
		self.setupPipeline()
	}

	self.audiosrc.Push(data)
}

func (self *WebRTCStreamer) PushVideoFrame(data []byte, timestamp uint) {
	// todo

	if self.pipeline == nil {
		self.setupPipeline()
	}

}
