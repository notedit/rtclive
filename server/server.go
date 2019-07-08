package server

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	mediaserver "github.com/notedit/media-server-go"
	"github.com/notedit/rtclive/config"
	"github.com/notedit/rtclive/router"
	"github.com/notedit/rtmp-lib"
)

const (
	rtmpproto   = "rtmp://"
	webrtcproto = "webrtc://"
)

type Server struct {
	sync.Mutex
	httpServer *gin.Engine

	cfg *config.Config

	rtmpServer *rtmp.Server

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

	//s.httpServer.POST("/api/publish", s.publish)
	//s.httpServer.POST("/api/unpublish", s.unpublish)

	s.httpServer.GET("/test", s.test)

	s.httpServer.POST("/api/play", s.play)
	s.httpServer.POST("/api/unplay", s.unplay)

	s.httpServer.POST("/api/relay", s.relay)

	address := ":" + strconv.Itoa(s.cfg.Server.Port)

	fmt.Println("start listen on " + address)

	if s.cfg.Rtmp != nil {
		go s.startRtmp()
	}

	s.httpServer.Run(":" + strconv.Itoa(s.cfg.Server.Port))
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

		if !s.cfg.Relay {
			c.JSON(200, gin.H{"s": 10002, "e": "does not exist"})
			return
		}

		parsedURL, err := url.Parse(data.StreamURL)

		if err != nil {
			c.JSON(200, gin.H{"s": 10004, "e": "stream url is invalid"})
			return
		}

		// todo rtmp or webrtc relay

		fmt.Println(data.StreamURL)
		fmt.Println(parsedURL.Path)

		conn, err := rtmp.Dial(data.StreamURL)

		if err != nil {
			c.JSON(200, gin.H{"s": 10003, "e": "stream relay error"})
			return
		}

		endpoint := s.getEndpoint(data.StreamID)
		mediarouter = router.NewMediaRouter(data.StreamID, endpoint, s.cfg.Capabilities, true)
		publisher := mediarouter.CreateRTMPPublisher(data.StreamID, conn)

		done := publisher.Start()

		s.addRouter(mediarouter)

		go func() {
			<-done
			fmt.Println("publisher done ")
			mediarouter.Stop()
			s.removeRouter(mediarouter.GetID())
		}()

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

func (s *Server) test(c *gin.Context) {
	c.String(200, "hello world")
}

func (s *Server) relay(c *gin.Context) {

}

func (s *Server) startRtmp() {

	s.rtmpServer = &rtmp.Server{
		Addr: fmt.Sprintf("%s:%d", s.cfg.Rtmp.Host, s.cfg.Rtmp.Port),
	}

	s.rtmpServer.HandlePublish = func(conn *rtmp.Conn) {

		streaminfo := strings.Split(conn.URL.Path, "/")

		if len(streaminfo) <= 2 {
			fmt.Println("rtmp url does not match, rtmp url should like rtmp://host:port/app/stream")
			conn.Close()
			return
		}

		streamID := streaminfo[len(streaminfo)-1]
		appName := streaminfo[len(streaminfo)-2]

		fmt.Printf("publishing stream %s in app %s\n", streamID, appName)

		endpoint := s.getEndpoint(streamID)
		capabilities := s.cfg.Capabilities

		mediarouter := router.NewMediaRouter(streamID, endpoint, capabilities, true)
		publisher := mediarouter.CreateRTMPPublisher(streamID, conn)
		s.addRouter(mediarouter)

		done := publisher.Start()

		err := <-done

		fmt.Println("error ", err)
		mediarouter.Stop()
		s.removeRouter(mediarouter.GetID())
	}

	s.rtmpServer.ListenAndServe()
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
