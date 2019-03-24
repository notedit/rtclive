package router

import (
	"sync"

	"github.com/gofrs/uuid"
	mediaserver "github.com/notedit/media-server-go"
	"github.com/notedit/media-server-go/sdp"
)

// Publisher interface
type Publisher interface {
	GetID() string
	GetVideoTrack() *mediaserver.IncomingStreamTrack
	GetAudioTrack() *mediaserver.IncomingStreamTrack
	Stop()
}

// Subscriber struct
type Subscriber struct {
	id          string
	publisherId string
	outgoing    *mediaserver.OutgoingStream
	transport   *mediaserver.Transport
}

// GetID get subscriber id
func (s *Subscriber) GetID() string {
	return s.id
}

// GetPublisherID get publisher id
func (s *Subscriber) GetPublisherID() string {
	return s.publisherId
}

// GetStream  get outgoing stream
func (s *Subscriber) GetStream() *mediaserver.OutgoingStream {
	return s.outgoing
}

// GetTransport transport
func (s *Subscriber) GetTransport() *mediaserver.Transport {
	return s.transport
}

// Stop stop it
func (s *Subscriber) Stop() {
	s.outgoing.Stop()
	s.transport.Stop()
}

// MediaRouter mediarouter
type MediaRouter struct {
	routerID     string
	capabilities map[string]*sdp.Capability
	endpoint     *mediaserver.Endpoint
	publisher    Publisher
	subscribers  map[string]*Subscriber
	originUrl    string
	origin       bool
	sync.Mutex
}

func NewMediaRouter(routerID string, endpoint *mediaserver.Endpoint, capabilities map[string]*sdp.Capability, origin bool) *MediaRouter {
	router := &MediaRouter{}
	router.routerID = routerID
	router.endpoint = endpoint
	router.capabilities = capabilities
	router.origin = origin

	router.subscribers = make(map[string]*Subscriber)
	return router
}

func (r *MediaRouter) GetID() string {
	return r.routerID
}

func (r *MediaRouter) IsOrgin() bool {
	return r.origin
}

func (r *MediaRouter) GetPublisher() Publisher {
	return r.publisher
}

func (r *MediaRouter) SetPublisher(publisher Publisher) {
	r.publisher = publisher
}

func (r *MediaRouter) SetOriginUrl(origin string) {
	r.originUrl = origin
}

func (s *MediaRouter) GetOriginUrl() string {
	return s.originUrl
}

func (s *MediaRouter) GetSubscribers() map[string]*Subscriber {
	return s.subscribers
}

func (r *MediaRouter) CreateRTCPublisher(sdpStr string) *RTCPublisher {

	publisher := NewRTCPublisher(sdpStr, r.endpoint, r.capabilities)
	r.publisher = publisher
	return publisher
}

func (r *MediaRouter) CreateRelayPublisher(offerStr string, answerStr string) *RTCPublisher {

	publisher := NewRelayPublisher(offerStr, answerStr, r.endpoint, r.capabilities)
	r.publisher = publisher
	return publisher

}

func (r *MediaRouter) CreateRTMPPublisher(streamId string) *RTMPPublisher {

	publisher := NewRTMPPublisher(streamId, r.capabilities)
	r.publisher = publisher

	return publisher
}

func (r *MediaRouter) CreateSubscriber(sdpStr string) (*Subscriber, string) {
	offer, err := sdp.Parse(sdpStr)
	if err != nil {
		panic(err)
	}

	transport := r.endpoint.CreateTransport(offer, nil)
	transport.SetRemoteProperties(offer.GetMedia("audio"), offer.GetMedia("video"))

	answer := offer.Answer(transport.GetLocalICEInfo(),
		transport.GetLocalDTLSInfo(),
		r.endpoint.GetLocalCandidates(),
		r.capabilities)

	transport.SetLocalProperties(answer.GetMedia("audio"), answer.GetMedia("video"))

	subId := uuid.Must(uuid.NewV4()).String()

	audio := r.publisher.GetAudioTrack() != nil
	video := r.publisher.GetVideoTrack() != nil

	outgoing := transport.CreateOutgoingStreamWithID(subId, audio, video)

	if audio {
		outgoing.GetAudioTracks()[0].AttachTo(r.publisher.GetAudioTrack())
	}

	if video {
		outgoing.GetVideoTracks()[0].AttachTo(r.publisher.GetVideoTrack())
	}

	subscriber := &Subscriber{
		id:        subId,
		outgoing:  outgoing,
		transport: transport,
	}

	r.Lock()
	r.subscribers[subId] = subscriber
	r.Unlock()

	answer.AddStream(outgoing.GetStreamInfo())

	return subscriber, answer.String()
}

func (r *MediaRouter) StopSubscriber(subscriberId string) {
	subscriber := r.subscribers[subscriberId]
	if subscriber == nil {
		return
	}

	subscriber.Stop()

	r.Lock()
	delete(r.subscribers, subscriberId)
	r.Unlock()
}

func (r *MediaRouter) Stop() {
	r.Lock()
	defer r.Unlock()
	if r.publisher != nil {
		r.publisher.Stop()
	}

	for _, subscriber := range r.subscribers {
		subscriber.Stop()
	}

	r.publisher = nil
	r.subscribers = nil
}
