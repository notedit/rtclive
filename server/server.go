package server

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req"
	mediaserver "github.com/notedit/media-server-go"
	"github.com/notedit/rtclive/config"
	"github.com/notedit/rtclive/router"
)

type Server struct {
	sync.Mutex
	httpServer *gin.Engine

	cfg *config.Config

	endpoints map[string]*mediaserver.Endpoint
	routers   map[string]*router.MediaRouter
}

func New(cfg *config.Config) *Server {

	server := &Server{}
	server.cfg = cfg

	gin.SetMode(gin.ReleaseMode)
	httpServer := gin.Default()
	httpServer.Use(cors.Default())

	server.httpServer = httpServer
	server.endpoints = make(map[string]*mediaserver.Endpoint)
	server.routers = make(map[string]*router.MediaRouter)

	return server
}

// ListenAndServe  start to listen and serve
func (s *Server) ListenAndServe() {

	// s.httpServer.POST("/pull", s.pullStream)
	// s.httpServer.POST("/unpull", s.unpullStream)

	s.httpServer.POST("/publish", s.publish)
	s.httpServer.POST("/unpublish", s.unpublish)
	s.httpServer.POST("/play", s.play)
	s.httpServer.POST("/unplay", s.unplay)

	s.httpServer.POST("/onrelay", s.onrelay)

	address := s.cfg.Server.Host + ":" + strconv.Itoa(s.cfg.Server.Port)

	fmt.Println("start listen on " + address)

	s.httpServer.Run(address)
}

func (s *Server) play(c *gin.Context) {

	var data struct {
		StreamURL string `json:"streamUrl"`
		StreamID  string `json:"streamId"`
		Sdp       string `json:"sdp"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10001, "e": err})
		return
	}

	mediarouter := s.getRouter(data.StreamID)

	if mediarouter == nil {

		if s.cfg.Relay == nil {
			c.JSON(200, gin.H{"s": 10002, "e": "does not exist"})
			return
		}
		// now we start to relay
		relayStreamURL, err := s.relayRequest(data.StreamID, data.StreamURL)

		fmt.Println(relayStreamURL)

		if err != nil {
			c.JSON(200, gin.H{"s": 10002, "e": "does not exist"})
			return
		}

		if strings.HasPrefix(relayStreamURL, "rtmp://") {
			// rtmp relay

			endpoint := s.getEndpoint(data.StreamID)
			mediarouter = router.NewMediaRouter(data.StreamID, endpoint, s.cfg.Capabilities, true)
			publisher := mediarouter.CreateRTMPPublisher(data.StreamID, relayStreamURL)

			done := publisher.Start()

			go func() {
				<-done
				fmt.Println("publisher done ")
				mediarouter.Stop()
			}()

			s.addRouter(mediarouter)

		} else if strings.HasPrefix(relayStreamURL, "webrtc://") {
			// webrtc relay
		}
	}

	subscriber := mediarouter.CreateSubscriber(data.Sdp)

	answer := subscriber.GetAnswer()

	c.JSON(200, gin.H{
		"s": 10000,
		"d": map[string]string{
			"sdp":          answer,
			"subscriberId": subscriber.GetID(),
		}})

}

func (s *Server) publish(c *gin.Context) {

	var data struct {
		StreamURL string `json:"streamUrl"`
		StreamID  string `json:"streamId"`
		Sdp       string `json:"sdp"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10001, "e": err})
		return
	}

	capabilities := s.cfg.Capabilities

	endpoint := s.getEndpoint(data.StreamID)

	mediarouter := router.NewMediaRouter(data.StreamID, endpoint, capabilities, true)
	publisher := mediarouter.CreatePublisher(data.Sdp)
	s.addRouter(mediarouter)

	answer := publisher.GetAnswer()

	fmt.Println("answer", answer)

	c.JSON(200, gin.H{
		"s": 10000,
		"d": map[string]string{
			"sdp": answer,
		}})
}

func (s *Server) unpublish(c *gin.Context) {

	var data struct {
		StreamURL string `json:"streamUrl"`
		StreamID  string `json:"streamId"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10001, "e": err})
		return
	}

	mediarouter := s.getRouter(data.StreamID)

	if mediarouter != nil {
		mediarouter.Stop()
		s.removeRouter(mediarouter.GetID())
	}

	c.JSON(200, gin.H{
		"s": 10000,
		"d": map[string]string{},
	})
}

func (s *Server) unplay(c *gin.Context) {

	var data struct {
		StreamURL    string `json:"streamUrl"`
		StreamID     string `json:"streamId"`
		SubscriberID string `json:"subscriberId"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10001, "e": err})
		return
	}

	mediarouter := s.getRouter(data.StreamID)

	if mediarouter != nil {
		mediarouter.StopSubscriber(data.SubscriberID)
	}

	c.JSON(200, gin.H{
		"s": 10000,
		"d": map[string]string{},
	})
}

func (s *Server) onrelay(c *gin.Context) {

	c.JSON(200, gin.H{
		"url": "rtmp://127.0.0.1/live/stream",
	})
}

func (s *Server) relayRequest(streamID string, requestStreamURL string) (streamURL string, err error) {

	res, err := req.Post(s.cfg.Relay.URL, req.BodyJSON(map[string]string{
		"streamId":  streamID,
		"streamUrl": requestStreamURL,
	}))

	if err != nil {
		panic(err)
	}

	var ret struct {
		URL string `json:"url"`
	}

	err = res.ToJSON(&ret)

	if err != nil {
		panic(err)
	}

	if !(strings.HasPrefix(ret.URL, "rtmp://") || strings.HasPrefix(ret.URL, "webrtc://")) {
		return "", errors.New("url error ")
	}

	return ret.URL, nil

}

func (s *Server) pullStream(c *gin.Context) {

	var data struct {
		StreamID string `json:"streamId"`
		Sdp      string `json:"sdp"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10001, "e": err})
		return
	}

	mediaRouter := s.getRouter(data.StreamID)

	if mediaRouter == nil {
		c.JSON(200, gin.H{"s": 10002, "e": "can not find stream"})
		return
	}

	subscriber := mediaRouter.CreateSubscriber(data.Sdp)

	fmt.Println("answer", subscriber.GetAnswer())

	c.JSON(200, gin.H{"s": 10000, "d": map[string]string{
		"sdp":          subscriber.GetAnswer(),
		"subscriberId": subscriber.GetID(),
	}})
}

func (s *Server) unpullStream(c *gin.Context) {

	var data struct {
		StreamID     string `json:"streamId"`
		SubscriberID string `json:"subscriberId"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10001, "e": err})
		return
	}

	mediaRouter := s.getRouter(data.StreamID)

	if mediaRouter == nil {
		c.JSON(200, gin.H{"s": 10002, "e": "can not find stream"})
		return
	}

	mediaRouter.StopSubscriber(data.SubscriberID)

	c.JSON(200, gin.H{"s": 10000, "d": map[string]string{}})
}

func (s *Server) getEndpoint(streamID string) *mediaserver.Endpoint {
	defer s.Unlock()
	s.Lock()

	if s.endpoints[streamID] != nil {
		return s.endpoints[streamID]
	}

	endpoint := mediaserver.NewEndpoint(s.cfg.Media.Endpoint)
	s.endpoints[streamID] = endpoint

	return endpoint
}

func (s *Server) removeEndpoint(streamID string) {
	defer s.Unlock()
	s.Lock()

	delete(s.endpoints, streamID)
}

func (s *Server) getRouter(routerID string) *router.MediaRouter {
	s.Lock()
	defer s.Unlock()
	return s.routers[routerID]
}

func (s *Server) addRouter(router *router.MediaRouter) {
	s.Lock()
	defer s.Unlock()
	s.routers[router.GetID()] = router
}

func (s *Server) removeRouter(routerID string) {
	s.Lock()
	defer s.Unlock()
	delete(s.routers, routerID)
}
