package router

import (
	mediaserver "github.com/notedit/media-server-go"
	"github.com/notedit/media-server-go/sdp"
)

// RTCPublisher struct
type RTCPublisher struct {
	id         string
	videotrack *mediaserver.IncomingStreamTrack
	audiotrack *mediaserver.IncomingStreamTrack
	transport  *mediaserver.Transport
	answer     string
}

// NewRTCPublisher create new rtc publisher
func NewRTCPublisher(sdpStr string, endpoint *mediaserver.Endpoint, capabilities map[string]*sdp.Capability) *RTCPublisher {

	offer, err := sdp.Parse(sdpStr)
	if err != nil {
		panic(err)
	}

	if offer.GetFirstStream() == nil {
		panic("can not find stream info")
	}

	transport := endpoint.CreateTransport(offer, nil)
	transport.SetRemoteProperties(offer.GetMedia("audio"), offer.GetMedia("video"))

	answerInfo := offer.Answer(transport.GetLocalICEInfo(),
		transport.GetLocalDTLSInfo(),
		endpoint.GetLocalCandidates(),
		capabilities)

	transport.SetLocalProperties(answerInfo.GetMedia("audio"), answerInfo.GetMedia("video"))

	streamInfo := offer.GetFirstStream()
	incoming := transport.CreateIncomingStream(streamInfo)

	var videoTrack *mediaserver.IncomingStreamTrack
	var audioTrack *mediaserver.IncomingStreamTrack

	if len(incoming.GetVideoTracks()) > 0 {
		videoTrack = incoming.GetVideoTracks()[0]
	}

	if len(incoming.GetAudioTracks()) > 0 {
		audioTrack = incoming.GetAudioTracks()[0]
	}

	publisher := &RTCPublisher{
		id:         incoming.GetID(),
		videotrack: videoTrack,
		audiotrack: audioTrack,
		transport:  transport,
		answer:     answerInfo.String(),
	}
	return publisher
}

func NewRelayPublisher(offerStr string, answerStr string, endpoint *mediaserver.Endpoint, capabilities map[string]*sdp.Capability) *RTCPublisher {

	offer, err := sdp.Parse(offerStr)
	if err != nil {
		panic(err)
	}

	answer, err := sdp.Parse(answerStr)
	if err != nil {
		panic(err)
	}

	if answer.GetFirstStream() == nil {
		panic("can not get stream info")
	}

	transport := endpoint.CreateTransport(answer, offer, true)

	transport.SetLocalProperties(offer.GetAudioMedia(), offer.GetVideoMedia())
	transport.SetRemoteProperties(answer.GetAudioMedia(), answer.GetVideoMedia())

	streamInfo := answer.GetFirstStream()

	incoming := transport.CreateIncomingStream(streamInfo)

	var videoTrack *mediaserver.IncomingStreamTrack
	var audioTrack *mediaserver.IncomingStreamTrack

	if len(incoming.GetVideoTracks()) > 0 {
		videoTrack = incoming.GetVideoTracks()[0]
	}

	if len(incoming.GetAudioTracks()) > 0 {
		audioTrack = incoming.GetAudioTracks()[0]
	}

	publisher := &RTCPublisher{
		id:         incoming.GetID(),
		videotrack: videoTrack,
		audiotrack: audioTrack,
		transport:  transport,
	}

	return publisher
}

// GetID  get publisher id
func (p *RTCPublisher) GetID() string {
	return p.id
}

// GetAnswer get answer str
func (p *RTCPublisher) GetAnswer() string {
	return p.answer
}

// GetVideoTrack  get video track
func (p *RTCPublisher) GetVideoTrack() *mediaserver.IncomingStreamTrack {
	return p.videotrack
}

// GetAudioTrack  get audio track
func (p *RTCPublisher) GetAudioTrack() *mediaserver.IncomingStreamTrack {
	return p.audiotrack
}

// Stop  stop this publisher
func (p *RTCPublisher) Stop() {

	if p.videotrack != nil {
		p.videotrack.Stop()
	}

	if p.audiotrack != nil {
		p.audiotrack.Stop()
	}

	if p.transport != nil {
		p.transport.Stop()
	}

}
