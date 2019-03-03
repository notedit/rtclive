package streamer

import (
	"github.com/hajimehoshi/oto"

	gstreamer "github.com/notedit/gstreamer-go"
	mediaserver "github.com/notedit/media-server-go"
	opus "gopkg.in/hraban/opus.v2"
)

//const rtp2rtmp = "appsrc is-live=true do-timestamp=true name=audiosrc ! opusparse ! opusdec ! audioconvert ! faac ! mux.  videotestsrc num-buffers=10000 do-timestamp=true ! video/x-raw,framerate=25/1 ! x264enc ! mux.  flvmux name=mux ! filesink location=test.flv"

const rtp2rtmp = "appsrc is-live=true do-timestamp=true name=audiosrc ! opusparse ! mux. appsrc is-live=true do-timestamp=true name=videosrc ! h264parse ! mux.  matroskamux name=mux ! filesink location=test.mkv"

type WebRTCStreamer struct {
	rtmpUrl string

	audioTrack *mediaserver.IncomingStreamTrack
	videoTrack *mediaserver.IncomingStreamTrack

	pipeline *gstreamer.Pipeline
	audiosrc *gstreamer.Element
	videosrc *gstreamer.Element

	decoder *opus.Decoder

	player *oto.Player
}

func NewWebRTCStreamer(rtmpUrl string, audioTrack *mediaserver.IncomingStreamTrack, videoTrack *mediaserver.IncomingStreamTrack) *WebRTCStreamer {

	streamer := &WebRTCStreamer{}
	streamer.rtmpUrl = rtmpUrl

	streamer.audioTrack = audioTrack
	streamer.videoTrack = videoTrack

	streamer.decoder, _ = opus.NewDecoder(48000, 2)

	streamer.player, _ = oto.NewPlayer(48000, 2, 2, 4096)

	return streamer
}

func (self *WebRTCStreamer) setupPipeline() {

	pipeline, err := gstreamer.New(rtp2rtmp)
	if err != nil {
		panic(err)
	}

	self.pipeline = pipeline
	self.audiosrc = pipeline.FindElement("audiosrc")
	self.videosrc = pipeline.FindElement("videosrc")

	self.pipeline.Start()

}

func (self *WebRTCStreamer) PushAudioFrame(data []byte, timestamp uint) {

	if self.pipeline == nil {
		self.setupPipeline()
	}

	// pcm := make([]int16, 4096)
	// frameSize, err := self.decoder.Decode(data, pcm)

	// if err != nil {
	// 	fmt.Println(err, frameSize)
	// }

	// fmt.Println(frameSize, timestamp)

	// pcm = pcm[:frameSize*2]

	// buf := new(bytes.Buffer)

	// err = binary.Write(buf, binary.LittleEndian, pcm)
	// if err != nil {
	// 	fmt.Println("binary.Write failed:", err)
	// }

	// fmt.Println(len(buf.Bytes()))

	// _, err = self.player.Write(buf.Bytes())

	// if err != nil {
	// 	fmt.Println(err)
	// }

	self.audiosrc.Push(data)

}

func (self *WebRTCStreamer) PushVideoFrame(data []byte, timestamp uint) {

	if self.pipeline == nil {
		self.setupPipeline()
	}

	self.videosrc.Push(data)
}
