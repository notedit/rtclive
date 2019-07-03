package router

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	mediaserver "github.com/notedit/media-server-go"
	"github.com/notedit/sdp"
)

// RTCSubscriber is a Subscriber interface
type RTCSubscriber struct {
	id          string
	publisherID string
	answer      string
	outgoing    *mediaserver.OutgoingStream
	transport   *mediaserver.Transport
	iceticker   *time.Ticker
	icestats    mediaserver.ICEStats
}

// NewRTCSubscriber create new subscriber
func NewRTCSubscriber(sdpStr string, endpoint *mediaserver.Endpoint, capabilities map[string]*sdp.Capability) *RTCSubscriber {

	offer, err := sdp.Parse(sdpStr)
	if err != nil {
		panic(err)
	}

	transport := endpoint.CreateTransport(offer, nil)
	transport.SetRemoteProperties(offer.GetMedia("audio"), offer.GetMedia("video"))

	answer := offer.Answer(transport.GetLocalICEInfo(),
		transport.GetLocalDTLSInfo(),
		endpoint.GetLocalCandidates(),
		capabilities)

	transport.SetLocalProperties(answer.GetMedia("audio"), answer.GetMedia("video"))

	subID := uuid.Must(uuid.NewV4()).String()

	outgoing := transport.CreateOutgoingStreamWithID(subID, true, true)

	answer.AddStream(outgoing.GetStreamInfo())

	subscriber := &RTCSubscriber{
		id:        subID,
		outgoing:  outgoing,
		transport: transport,
		answer:    answer.String(),
	}

	subscriber.iceticker = time.NewTicker(5 * time.Second)

	go subscriber.runIceTicker()

	return subscriber
}

// GetID get subscriber id
func (s *RTCSubscriber) GetID() string {
	return s.id
}

// GetPublisherID get publisher id
func (s *RTCSubscriber) GetPublisherID() string {
	return s.publisherID
}

// Attach to a publisher
func (s *RTCSubscriber) Attach(publisher Publisher) {

	if publisher.GetAudioTrack() != nil {
		s.outgoing.GetAudioTracks()[0].AttachTo(publisher.GetAudioTrack())
	} else {
		fmt.Println("Attach audio track")
	}

	if publisher.GetVideoTrack() != nil {
		s.outgoing.GetVideoTracks()[0].AttachTo(publisher.GetVideoTrack())
	} else {
		fmt.Println("Attach video track")
	}
}

// GetTransport transport
func (s *RTCSubscriber) GetTransport() *mediaserver.Transport {
	return s.transport
}

// GetAnswer return the answer sdp
func (s *RTCSubscriber) GetAnswer() string {
	return s.answer
}

// Stop stop it
func (s *RTCSubscriber) Stop() {

	s.outgoing.Stop()
	s.transport.Stop()

	s.iceticker.Stop()
}

func (s *RTCSubscriber) runIceTicker() {

	for _ = range s.iceticker.C {
		icestats := s.transport.GetICEStats()
		fmt.Printf("Old RequestsReceived %d, New RequestsReceived %d\n", s.icestats.RequestsReceived, icestats.RequestsReceived)
	}
}
