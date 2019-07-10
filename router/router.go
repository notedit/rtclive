package router

import (
	"sync"

	mediaserver "github.com/notedit/media-server-go"
	"github.com/notedit/sdp"
)

// Publisher interface
type Publisher interface {
	GetID() string
	GetAnswer() string
	GetVideoTrack() *mediaserver.IncomingStreamTrack
	GetAudioTrack() *mediaserver.IncomingStreamTrack
	Stop()
}

// Subscriber interface
type Subscriber interface {
	GetID() string
	GetAnswer() string
	Attach(publisher Publisher)
	GetTransport() *mediaserver.Transport
	Stop()
}

// MediaRouter mediarouter
type MediaRouter struct {
	routerID     string
	capabilities map[string]*sdp.Capability
	endpoint     *mediaserver.Endpoint
	publisher    Publisher
	subscribers  map[string]Subscriber
	origin       bool
	sync.Mutex
}

func NewMediaRouter(routerID string, endpoint *mediaserver.Endpoint, capabilities map[string]*sdp.Capability, origin bool) *MediaRouter {
	router := &MediaRouter{}
	router.routerID = routerID
	router.endpoint = endpoint
	router.capabilities = capabilities
	router.origin = origin

	router.subscribers = make(map[string]Subscriber)
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

func (s *MediaRouter) GetSubscribers() map[string]Subscriber {
	return s.subscribers
}

func (s *MediaRouter) GetSubscribersCount() int {
	return len(s.subscribers)
}

func (r *MediaRouter) CreatePublisher(sdpStr string) *RTCPublisher {

	publisher := NewRTCPublisher(sdpStr, r.endpoint, r.capabilities)
	r.publisher = publisher
	return publisher
}

func (r *MediaRouter) CreateRelayPublisher(offerStr string, answerStr string) *RTCPublisher {

	publisher := NewRelayPublisher(offerStr, answerStr, r.endpoint, r.capabilities)
	r.publisher = publisher
	return publisher
}

func (r *MediaRouter) CreateFFPublisher(streamID string, streamURL string) *FFPublisher {

	publisher := NewFFPublisher(streamID, streamURL, r.capabilities)
	r.publisher = publisher
	return publisher
}

func (r *MediaRouter) CreateSubscriber(sdpStr string) Subscriber {

	subscriber := NewRTCSubscriber(sdpStr, r.endpoint, r.capabilities)

	r.Lock()
	r.subscribers[subscriber.GetID()] = subscriber
	r.Unlock()

	subscriber.Attach(r.publisher)

	return subscriber
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

	if r.publisher != nil {
		r.publisher.Stop()
	}

	for _, subscriber := range r.subscribers {
		subscriber.Stop()
	}

	r.publisher = nil
	r.subscribers = nil
}
