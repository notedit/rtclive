package router

import (
	"bytes"
	"fmt"

	rtmp "github.com/notedit/rtmp-lib"
	"github.com/notedit/rtmp-lib/aac"
	"github.com/notedit/rtmp-lib/av"
	"github.com/notedit/rtmp-lib/h264"
	"github.com/notedit/sdp"

	gstreamer "github.com/notedit/gstreamer-go"
	mediaserver "github.com/notedit/media-server-go"
)

const video2rtp = `appsrc do-timestamp=true is-live=true name=videosrc ! h264parse ! rtph264pay timestamp-offset=0 config-interval=-1 pt=%d ! udpsink host=127.0.0.1 port=%d`
const audio2rtp = `appsrc is-live=true name=audiosrc ! aacparse ! faad ! audioconvert ! audioresample ! audio/x-raw,rate=48000 ! opusenc ! rtpopuspay timestamp-offset=0 pt=%d ! udpsink host=127.0.0.1 port=%d`

const audio2rtp2 = `appsrc is-live=true name=audiosrc ! aacparse ! faad ! audioconvert ! audioresample ! audio/x-raw,rate=48000,channels=1 ! autoaudiosink`

var startCodeBytes = []byte{0, 0, 0, 1}

// RTMPPublisher  rtmp publisher
type RTMPPublisher struct {
	id string

	videoPipeline *gstreamer.Pipeline
	audioPipeline *gstreamer.Pipeline

	streams        []av.CodecData
	videoCodecData h264.CodecData
	audioCodecData aac.CodecData

	adts []byte

	audiosrc *gstreamer.Element
	videosrc *gstreamer.Element

	videoSession *mediaserver.StreamerSession
	audioSession *mediaserver.StreamerSession

	videoWriteBuffer bytes.Buffer
	audioWriteBuffer bytes.Buffer

	capabilities map[string]*sdp.Capability

	rtmpURL string
	conn    *rtmp.Conn
}

// NewRTMPPublisher  new rtmp publisher
func NewRTMPPublisher(streamID string, conn *rtmp.Conn, capabilities map[string]*sdp.Capability) *RTMPPublisher {

	publisher := &RTMPPublisher{}
	publisher.id = streamID
	publisher.capabilities = capabilities
	publisher.conn = conn

	return publisher
}

// Start start the pipeline
func (p *RTMPPublisher) Start() <-chan error {

	done := make(chan error, 1)

	streams, err := p.conn.Streams()

	if err != nil {
		done <- err
		return done
	}

	for _, stream := range streams {
		if stream.Type() == av.H264 {
			p.videoCodecData = stream.(h264.CodecData)
		}
		if stream.Type() == av.AAC {
			p.audioCodecData = stream.(aac.CodecData)
			p.adts = make([]byte, 7)
		}
	}

	p.streams = streams

	videoMediaInfo := sdp.MediaInfoCreate("video", p.capabilities["video"])
	videoPt := videoMediaInfo.GetCodec("h264").GetType()
	p.videoSession = mediaserver.NewStreamerSession(videoMediaInfo)

	audioMediaInfo := sdp.MediaInfoCreate("audio", p.capabilities["audio"])
	audioPt := audioMediaInfo.GetCodec("opus").GetType()
	p.audioSession = mediaserver.NewStreamerSession(audioMediaInfo)

	video2rtpstr := fmt.Sprintf(video2rtp, videoPt, p.videoSession.GetLocalPort())
	fmt.Println(video2rtpstr)
	p.videoPipeline, err = gstreamer.New(video2rtpstr)
	if err != nil {
		panic(err)
	}
	p.videosrc = p.videoPipeline.FindElement("videosrc")

	audio2rtpstr := fmt.Sprintf(audio2rtp, audioPt, p.audioSession.GetLocalPort())

	//audio2rtpstr = audio2rtp2

	fmt.Println(audio2rtpstr)
	p.audioPipeline, err = gstreamer.New(audio2rtpstr)
	if err != nil {
		panic(err)
	}
	p.audiosrc = p.audioPipeline.FindElement("audiosrc")

	p.audioPipeline.Start()
	p.videoPipeline.Start()

	go p.handleMediaPacket(done)

	return done
}

// GetID  get publisher id
func (p *RTMPPublisher) GetID() string {
	return p.id
}

// GetAnswer get answer str
func (p *RTMPPublisher) GetAnswer() string {
	return ""
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

func (p *RTMPPublisher) handleMediaPacket(done chan error) {

	for {
		packet, err := p.conn.ReadPacket()
		//fmt.Println("got pakcet")
		if err != nil {
			done <- err
			break
		}

		p.writePacket(packet)
	}

}

func (p *RTMPPublisher) writePacket(packet av.Packet) {

	stream := p.streams[packet.Idx]

	if stream.Type() == av.H264 {

		if packet.IsKeyFrame {
			p.videoWriteBuffer.Write(startCodeBytes)
			p.videoWriteBuffer.Write(p.videoCodecData.SPS())
			p.videoWriteBuffer.Write(startCodeBytes)
			p.videoWriteBuffer.Write(p.videoCodecData.PPS())
			p.videosrc.Push(p.videoWriteBuffer.Bytes())
			p.videoWriteBuffer.Reset()
		}

		pktnalus, _ := h264.SplitNALUs(packet.Data)
		if len(pktnalus) > 1 {
			//fmt.Println("av.Packet has more than one nals")

		}
		for _, nalu := range pktnalus {
			p.videoWriteBuffer.Write(startCodeBytes)
			p.videoWriteBuffer.Write(nalu)
			p.videosrc.Push(p.videoWriteBuffer.Bytes())
			p.videoWriteBuffer.Reset()
		}
	}

	if stream.Type() == av.AAC {
		aac.FillADTSHeader(p.adts, p.audioCodecData.Config, 1024, len(packet.Data))
		p.audioWriteBuffer.Write(p.adts)
		p.audioWriteBuffer.Write(packet.Data)
		p.audiosrc.Push(p.audioWriteBuffer.Bytes())
		p.audioWriteBuffer.Reset()
	}
}

// Stop  stop this publisher
func (p *RTMPPublisher) Stop() {

	if p.audioSession != nil {
		p.audioSession.Stop()
	}

	if p.videoSession != nil {
		p.videoSession.Stop()
	}

	if p.videosrc != nil {
		p.videosrc.Stop()
	}

	if p.audiosrc != nil {
		p.audiosrc.Stop()
	}

	if p.videoPipeline != nil {
		p.videoPipeline.Stop()
	}

	if p.audioPipeline != nil {
		p.audioPipeline.Stop()
	}

}
