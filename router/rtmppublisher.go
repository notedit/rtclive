package router

import (
	"bytes"
	"fmt"

	"github.com/notedit/rtmp-lib/aac"
	"github.com/notedit/rtmp-lib/av"
	"github.com/notedit/rtmp-lib/h264"
	"github.com/notedit/sdp"

	gstreamer "github.com/notedit/gstreamer-go"
	mediaserver "github.com/notedit/media-server-go"
)

const audio2rtp = "appsrc do-timestamp=true is-live=true name=appsrc ! decodebin ! audioconvert ! audioresample ! opusenc ! rtpopuspay timestamp-offset=0 pt=%d ! appsink name=appsink"
const video2rtp = "appsrc do-timestamp=true is-live=true name=appsrc ! h264parse ! rtph264pay timestamp-offset=0 config-interval=-1 pt=%d ! appsink name=appsink"

// RTMPPublisher  rtmp publisher
type RTMPPublisher struct {
	id             string
	streams        []av.CodecData
	videoCodecData h264.CodecData
	audioCodecData aac.CodecData
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

// NewRTMPPublisher  new rtmp publisher
func NewRTMPPublisher(streamID string, capabilities map[string]*sdp.Capability) *RTMPPublisher {

	publisher := &RTMPPublisher{}
	publisher.id = streamID
	publisher.audioCapability = capabilities["audio"]
	publisher.videoCapability = capabilities["video"]
	return publisher
}

// WriteHeader  write header
func (p *RTMPPublisher) WriteHeader(streams []av.CodecData) error {

	p.streams = streams
	p.streamer = mediaserver.NewRawStreamer()

	for _, stream := range streams {
		if stream.Type() == av.H264 {

			h264Codec := stream.(h264.CodecData)
			p.videoCodecData = h264Codec

			videoMediaInfo := sdp.MediaInfoCreate("video", p.videoCapability)

			p.videoSession = p.streamer.CreateSession(videoMediaInfo)

			video2rtpstr := fmt.Sprintf(video2rtp, videoMediaInfo.GetCodec("h264").GetType())
			videoPipeline, err := gstreamer.New(video2rtpstr)
			if err != nil {
				panic(err)
			}

			p.videoPipeline = videoPipeline
			p.videosrc = videoPipeline.FindElement("appsrc")
			p.videosink = videoPipeline.FindElement("appsink")
			videoPipeline.Start()
			p.videoout = p.videosink.Poll()

			go func() {
				for rtp := range p.videoout {
					p.videoSession.Push(rtp)
				}
			}()
		}

		if stream.Type() == av.AAC {
			aacCodec := stream.(aac.CodecData)
			p.audioCodecData = aacCodec

			audioMediaInfo := sdp.MediaInfoCreate("audio", p.audioCapability)

			audio2rtpstr := fmt.Sprintf(audio2rtp, audioMediaInfo.GetCodec("opus").GetType())
			audioPipeline, err := gstreamer.New(audio2rtpstr)
			if err != nil {
				panic(err)
			}

			p.adtsheader = make([]byte, 7)

			p.audioPipeline = audioPipeline
			p.audiosrc = audioPipeline.FindElement("appsrc")
			p.audiosink = audioPipeline.FindElement("appsink")
			audioPipeline.Start()
			p.audioout = p.audiosink.Poll()

			p.audioSession = p.streamer.CreateSession(audioMediaInfo)

			go func() {
				for rtp := range p.audioout {
					fmt.Println("audio===", len(rtp))
					p.audioSession.Push(rtp)
				}
			}()
		}
	}

	return nil
}

// WritePacket  write packet
func (p *RTMPPublisher) WritePacket(packet av.Packet) error {

	stream := p.streams[packet.Idx]

	if stream.Type() == av.H264 {
		nalus := [][]byte{}

		if packet.IsKeyFrame {
			nalus = append(nalus, p.videoCodecData.SPS())
			nalus = append(nalus, p.videoCodecData.PPS())
		}

		if !p.spspps {
			p.videoWriteBuffer.Write([]byte{0, 0, 0, 1})
			p.videoWriteBuffer.Write(p.videoCodecData.SPS())
			p.videoWriteBuffer.Write([]byte{0, 0, 0, 1})
			p.videoWriteBuffer.Write(p.videoCodecData.PPS())
			p.videosrc.Push(p.videoWriteBuffer.Bytes())
			p.videoWriteBuffer.Reset()
			p.spspps = true
		}

		pktnalus, _ := h264.SplitNALUs(packet.Data)
		for _, nalu := range pktnalus {
			p.videoWriteBuffer.Write([]byte{0, 0, 0, 1})
			p.videoWriteBuffer.Write(nalu)
			p.videosrc.Push(p.videoWriteBuffer.Bytes())
			p.videoWriteBuffer.Reset()
		}
	}

	if stream.Type() == av.AAC {
		aac.FillADTSHeader(p.adtsheader, p.audioCodecData.Config, 1024, len(packet.Data))
		p.audioWriteBuffer.Write(p.adtsheader)
		p.audioWriteBuffer.Write(packet.Data)
		p.audiosrc.Push(p.audioWriteBuffer.Bytes())
		p.audioWriteBuffer.Reset()
	}

	return nil
}

// WriteTrailer  writer trailer
func (p *RTMPPublisher) WriteTrailer() error {
	return nil
}

// GetID  get publisher id
func (p *RTMPPublisher) GetID() string {
	return p.id
}

// GetVideoTrack get video track
func (p *RTMPPublisher) GetVideoTrack() *mediaserver.IncomingStreamTrack {

	if p.videoSession != nil {
		return p.videoSession.GetIncomingStreamTrack()
	}
	return nil
}

// GetAudioTrack get audio track
func (p *RTMPPublisher) GetAudioTrack() *mediaserver.IncomingStreamTrack {

	if p.audioSession != nil {
		return p.audioSession.GetIncomingStreamTrack()
	}
	return nil
}

// Stop  stop this publisher
func (p *RTMPPublisher) Stop() {

	if p.audioSession != nil {
		p.audioSession.Stop()
	}

	if p.videoSession != nil {
		p.videoSession.Stop()
	}

}
