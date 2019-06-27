package router

import (
	"fmt"

	"github.com/notedit/sdp"

	gstreamer "github.com/notedit/gstreamer-go"
	mediaserver "github.com/notedit/media-server-go"
)

const rtmp2rtp = `rtmpsrc location=%s ! flvdemux name=demux  
demux.audio_0 ! queue ! decodebin ! audioconvert ! audioresample ! opusenc ! rtpopuspay timestamp-offset=0 pt=%d ! appsink name=audiosink  
demux.video_0 ! queue ! h264parse ! rtph264pay timestamp-offset=0 config-interval=-1 pt=%d ! appsink name=videosink`

// RTMPPublisher  rtmp publisher
type RTMPPublisher struct {
	id string

	pipeline  *gstreamer.Pipeline
	audiosink *gstreamer.Element
	videosink *gstreamer.Element

	streamer     *mediaserver.RawStreamer
	videoSession *mediaserver.RawStreamerSession
	audioSession *mediaserver.RawStreamerSession

	rtmpURL string
}

// NewRTMPPublisher  new rtmp publisher
func NewRTMPPublisher(streamID string, rtmpURL string, capabilities map[string]*sdp.Capability) *RTMPPublisher {

	publisher := &RTMPPublisher{}
	publisher.id = streamID
	publisher.rtmpURL = rtmpURL

	publisher.streamer = mediaserver.NewRawStreamer()
	videoMediaInfo := sdp.MediaInfoCreate("video", capabilities["video"])
	videoPt := videoMediaInfo.GetCodec("h264").GetType()
	publisher.videoSession = publisher.streamer.CreateSession(videoMediaInfo)

	audioMediaInfo := sdp.MediaInfoCreate("audio", capabilities["audio"])
	audioPt := audioMediaInfo.GetCodec("opus").GetType()
	publisher.audioSession = publisher.streamer.CreateSession(audioMediaInfo)

	fmt.Println("videoPt", videoPt)
	fmt.Println("AudioPt", audioPt)

	pipelineStr := fmt.Sprintf(rtmp2rtp, rtmpURL, videoPt, audioPt)

	fmt.Println(pipelineStr)

	pipeline, err := gstreamer.New(pipelineStr)
	if err != nil {
		panic(err)
	}

	publisher.audiosink = pipeline.FindElement("audiosink")
	publisher.videosink = pipeline.FindElement("videosink")

	publisher.pipeline = pipeline

	return publisher
}

func (p *RTMPPublisher) Start() <-chan struct{} {

	done := make(chan struct{})

	videoout := p.videosink.Poll()

	go func() {
		for rtp := range videoout {
			p.videoSession.Push(rtp)
		}
	}()

	audioout := p.audiosink.Poll()

	go func() {
		for rtp := range audioout {
			p.audioSession.Push(rtp)
		}
	}()

	messages := p.pipeline.PullMessage()

	go func() {
		for message := range messages {
			if message.GetType() == gstreamer.MESSAGE_EOS || message.GetType() == gstreamer.MESSAGE_ERROR {
				done <- struct{}{}
			}
			fmt.Println(message.GetTypeName())
		}
	}()

	p.pipeline.Start()

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

// Stop  stop this publisher
func (p *RTMPPublisher) Stop() {

	if p.audioSession != nil {
		p.audioSession.Stop()
	}

	if p.videoSession != nil {
		p.videoSession.Stop()
	}

	if p.videosink != nil {
		p.videosink.Stop()
	}

	if p.audiosink != nil {
		p.audiosink.Stop()
	}

	if p.pipeline != nil {
		p.pipeline.Stop()
	}
}
