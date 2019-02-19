package rtmpstreamer

import (
	"fmt"

	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/codec/aacparser"
	"github.com/nareix/joy4/codec/h264parser"
	gstreamer "github.com/notedit/gstreamer-go"
	mediaserver "github.com/notedit/media-server-go"
	"github.com/notedit/media-server-go/sdp"
)

var audio2rtp = "appsrc is-live=true do-timestamp=true name=appsrc ! faad ! audioconvert ! audioresample ! audio/x-raw,rate=48000 ! opusenc ! rtpopuspay pt=%d ! appsink name=appsink"
var video2rtp = "appsrc is-live=true do-timestamp=true name=appsrc ! h264parse ! video/x-h264,stream-format=(string)byte-stream ! rtph264pay pt=%d ! appsink name=appsink"

type RtmpStreamer struct {
	id             string
	streams        []av.CodecData
	videoCodecData h264parser.CodecData
	audioCodecData aacparser.CodecData
	audioPipeline  *gstreamer.Pipeline
	videoPipeline  *gstreamer.Pipeline
	audiosink      *gstreamer.Element
	videosink      *gstreamer.Element
	audiosrc       *gstreamer.Element
	videosrc       *gstreamer.Element

	videoout <-chan []byte
	audioout <-chan []byte

	adtsheader []byte

	streamer        *mediaserver.RawStreamer
	videoSession    *mediaserver.RawStreamerSession
	audioSession    *mediaserver.RawStreamerSession
	audioCapability *sdp.Capability
	videoCapability *sdp.Capability
}

func NewRtmpStreamer(streamID string, audio *sdp.Capability, video *sdp.Capability) *RtmpStreamer {
	streamer := &RtmpStreamer{}
	streamer.audioCapability = audio
	streamer.videoCapability = video
	return streamer
}

// WriteHeader got sps and pps
func (self *RtmpStreamer) WriteHeader(streams []av.CodecData) error {

	self.streams = streams

	self.streamer = mediaserver.NewRawStreamer()

	for _, stream := range streams {
		if stream.Type() == av.H264 {
			h264Codec := stream.(h264parser.CodecData)
			self.videoCodecData = h264Codec

			videoMediaInfo := sdp.MediaInfoCreate("video", self.videoCapability)

			self.videoSession = self.streamer.CreateSession(videoMediaInfo)

			video2rtpstr := fmt.Sprintf(video2rtp, videoMediaInfo.GetCodec("h264").GetType())
			videoPipeline, err := gstreamer.New(video2rtpstr)
			if err != nil {
				panic(err)
			}

			self.videoPipeline = videoPipeline
			self.videosrc = videoPipeline.FindElement("appsrc")
			self.videosink = videoPipeline.FindElement("appsink")
			videoPipeline.Start()
			self.videoout = self.videosink.Poll()

			go func() {
				for {
					rtp, ok := <-self.videoout
					if !ok {
						break
					}
					self.videoSession.Push(rtp)
				}
			}()

		}
		if stream.Type() == av.AAC {
			aacCodec := stream.(aacparser.CodecData)
			self.audioCodecData = aacCodec

			audioMediaInfo := sdp.MediaInfoCreate("audio", self.audioCapability)

			audio2rtpstr := fmt.Sprintf(audio2rtp, audioMediaInfo.GetCodec("opus").GetType())
			audioPipeline, err := gstreamer.New(audio2rtpstr)
			if err != nil {
				panic(err)
			}

			self.adtsheader = make([]byte, 7)

			self.audioPipeline = audioPipeline
			self.audiosrc = audioPipeline.FindElement("appsrc")
			self.audiosink = audioPipeline.FindElement("appsink")
			audioPipeline.Start()
			self.audioout = self.audiosink.Poll()

			self.audioSession = self.streamer.CreateSession(audioMediaInfo)

			go func() {
				for {
					rtp, ok := <-self.audioout
					if !ok {
						break
					}
					self.audioSession.Push(rtp)
				}
			}()

		}
	}

	return nil
}

// WritePacket
func (self *RtmpStreamer) WritePacket(packet av.Packet) error {

	stream := self.streams[packet.Idx]

	if stream.Type() == av.H264 {
		nalus := [][]byte{}

		if packet.IsKeyFrame {
			nalus = append(nalus, self.videoCodecData.SPS())
			nalus = append(nalus, self.videoCodecData.PPS())
		}

		pktnalus, _ := h264parser.SplitNALUs(packet.Data)
		for _, nalu := range pktnalus {
			nalus = append(nalus, nalu)
		}

		for _, nalu := range nalus {
			naluc := []byte{0, 0, 0, 1}
			naluc = append(naluc, nalu...)
			self.videosrc.Push(naluc)
		}
	}

	if stream.Type() == av.AAC {

		adtsbuffer := []byte{}
		aacparser.FillADTSHeader(self.adtsheader, self.audioCodecData.Config, 1024, len(packet.Data))
		adtsbuffer = append(adtsbuffer, self.adtsheader...)
		adtsbuffer = append(adtsbuffer, packet.Data...)

		self.audiosrc.Push(adtsbuffer)
	}

	return nil
}

// WriteTrailer
func (self *RtmpStreamer) WriteTrailer() error {
	return nil
}

func (self *RtmpStreamer) HasVideo() bool {
	if self.videoPipeline != nil {
		return true
	}
	return false
}

func (self *RtmpStreamer) HasAudio() bool {
	if self.videoPipeline != nil {
		return true
	}
	return false
}

func (self *RtmpStreamer) GetVideoTrack() *mediaserver.IncomingStreamTrack {

	if self.videoSession != nil {
		return self.videoSession.GetIncomingStreamTrack()
	}
	return nil
}

func (self *RtmpStreamer) GetAuidoTrack() *mediaserver.IncomingStreamTrack {

	if self.audioSession != nil {
		return self.audioSession.GetIncomingStreamTrack()
	}
	return nil
}
