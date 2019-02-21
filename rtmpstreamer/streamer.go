package rtmpstreamer

import (
	"bytes"
	"fmt"

	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/codec/aacparser"
	"github.com/nareix/joy4/codec/h264parser"
	gstreamer "github.com/notedit/gstreamer-go"
	mediaserver "github.com/notedit/media-server-go"
	"github.com/notedit/media-server-go/sdp"
)

var audio2rtp = "appsrc do-timestamp=true is-live=true name=appsrc ! decodebin ! audioconvert ! audioresample ! opusenc ! rtpopuspay timestamp-offset=0 pt=%d ! appsink name=appsink"
var video2rtp = "appsrc do-timestamp=true is-live=true name=appsrc ! h264parse ! rtph264pay timestamp-offset=0 config-interval=-1 pt=%d ! appsink name=appsink"

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

	spspps bool

	videoWriteBuffer bytes.Buffer
	audioWriteBuffer bytes.Buffer

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
				for rtp := range self.videoout {
					fmt.Println("video===", len(rtp))
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
				for rtp := range self.audioout {
					fmt.Println("audio===", len(rtp))
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

		if !self.spspps {
			self.videoWriteBuffer.Write([]byte{0, 0, 0, 1})
			self.videoWriteBuffer.Write(self.videoCodecData.SPS())
			self.videoWriteBuffer.Write([]byte{0, 0, 0, 1})
			self.videoWriteBuffer.Write(self.videoCodecData.PPS())
			self.videosrc.Push(self.videoWriteBuffer.Bytes())
			self.videoWriteBuffer.Reset()
			self.spspps = true
		}

		pktnalus, _ := h264parser.SplitNALUs(packet.Data)
		for _, nalu := range pktnalus {
			self.videoWriteBuffer.Write([]byte{0, 0, 0, 1})
			self.videoWriteBuffer.Write(nalu)
			self.videosrc.Push(self.videoWriteBuffer.Bytes())
			self.videoWriteBuffer.Reset()
		}
	}

	if stream.Type() == av.AAC {

		aacparser.FillADTSHeader(self.adtsheader, self.audioCodecData.Config, 1024, len(packet.Data))
		self.audioWriteBuffer.Write(self.adtsheader)
		self.audioWriteBuffer.Write(packet.Data)
		self.audiosrc.Push(self.audioWriteBuffer.Bytes())
		self.audioWriteBuffer.Reset()
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
