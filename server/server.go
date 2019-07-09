package server

import (
	"fmt"
	"log"
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
	"github.com/notedit/rtmp-lib/av"
	"github.com/notedit/rtmp-lib/pubsub"
)

const (
	rtmpproto   = "rtmp://"
	webrtcproto = "webrtc://"
)

type Channel struct {
	que *pubsub.Queue
}

type Server struct {
	sync.RWMutex
	httpServer *gin.Engine

	cfg *config.Config

	rtmpChannels map[string]*Channel
	rtmpServer   *rtmp.Server

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
	server.rtmpChannels = make(map[string]*Channel)
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

	parsedURL, err := url.Parse(data.StreamURL)
	if err != nil {
		c.JSON(200, gin.H{"s": 10004, "e": "stream url is invalid"})
		return
	}

	mediarouter := s.getRouter(data.StreamID)

	if mediarouter == nil {

		var relayStreamURL string
		// this is a rtmp push stream, we relay it from local
		if s.getChannel(data.StreamID) != nil {
			streaminfo := strings.Split(parsedURL.Path, "/")
			if len(streaminfo) <= 2 {
				fmt.Println("rtmp url does not match, rtmp url should like rtmp://host:port/app/stream")
				return
			}
			streamID := streaminfo[len(streaminfo)-1]
			appName := streaminfo[len(streaminfo)-2]
			relayStreamURL = fmt.Sprintf("rtmp://localhost:%d/%s/%s", s.cfg.Rtmp.Port, appName, streamID)
		} else {
			relayStreamURL = data.StreamURL
		}

		endpoint := s.getEndpoint(data.StreamID)
		mediarouter = router.NewMediaRouter(data.StreamID, endpoint, s.cfg.Capabilities, true)
		publisher := mediarouter.CreateFFPublisher(data.StreamID, relayStreamURL)

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

	s.rtmpServer.HandlePlay = func(conn *rtmp.Conn) {

		streaminfo := strings.Split(conn.URL.Path, "/")

		if len(streaminfo) <= 2 {
			fmt.Println("rtmp url does not match, rtmp url should like rtmp://host:port/app/stream")
			conn.Close()
			return
		}

		streamID := streaminfo[len(streaminfo)-1]
		appName := streaminfo[len(streaminfo)-2]

		fmt.Printf("publishing stream %s in app %s\n", streamID, appName)

		ch := s.getChannel(streamID)

		if ch != nil {
			cursor := ch.que.Latest()
			streams, err := cursor.Streams()
			if err != nil {
				panic(err)
			}
			conn.WriteHeader(streams)
			for {
				packet, err := cursor.ReadPacket()
				if err != nil {
					break
				}
				err = conn.WritePacket(packet)
				if err != nil {
					break
				}
			}
		}
		conn.Close()
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

		ch := &Channel{}
		ch.que = pubsub.NewQueue()
		ch.que.SetMaxGopCount(0)

		s.addChannel(streamID, ch)

		var streams []av.CodecData
		var err error

		if streams, err = conn.Streams(); err != nil {
			fmt.Println(err)
		} else {
			ch.que.WriteHeader(streams)
			for {
				var pkt av.Packet
				if pkt, err = conn.ReadPacket(); err != nil {
					break
				}
				ch.que.WritePacket(pkt)
			}
		}
		s.removeChannel(streamID)
		ch.que.Close()
	}

	err := s.rtmpServer.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}
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

func (s *Server) addChannel(streamID string, channel *Channel) {
	s.Lock()
	defer s.Unlock()
	s.rtmpChannels[streamID] = channel
}

func (s *Server) getChannel(streamID string) *Channel {
	s.RLock()
	defer s.RUnlock()
	return s.rtmpChannels[streamID]
}

func (s *Server) removeChannel(streamID string) {
	s.Lock()
	defer s.Unlock()
	delete(s.rtmpChannels, streamID)
}
