package server

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	mediaserver "github.com/notedit/media-server-go"
	"github.com/notedit/rtclive/config"
	"github.com/notedit/rtclive/router"
	"github.com/notedit/rtclive/store"
)

type Server struct {
	httpServer *gin.Engine
	endpoint   *mediaserver.Endpoint
	cfg        *config.Config
}

func New(cfg *config.Config) *Server {

	server := &Server{}
	server.cfg = cfg

	gin.SetMode(gin.ReleaseMode)
	httpServer := gin.Default()
	httpServer.Use(cors.Default())

	server.httpServer = httpServer
	server.endpoint = mediaserver.NewEndpoint(cfg.Media.Endpoint)

	return server
}

func (s *Server) ListenAndServe() {

	// s.httpServer.POST("/pull", s.pullStream)
	// s.httpServer.POST("/unpull", s.unpullStream)

	s.httpServer.POST("/publish", s.publish)
	s.httpServer.POST("/unpublish", s.unpublish)
	s.httpServer.POST("/play", s.play)
	s.httpServer.POST("/unplay", s.unplay)

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

	router := store.GetRouter(data.StreamID)

	if router == nil {
		c.JSON(200, gin.H{"s": 10002, "e": "does not exist"})
		return
	}

	subscriber := router.CreateSubscriber(data.Sdp)

	answer := subscriber.GetAnswer()

	fmt.Println("answer", answer)

	c.JSON(200, gin.H{
		"s": 10000,
		"d": map[string]string{
			"sdp": answer,
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

	mediarouter := router.NewMediaRouter(data.StreamID, s.endpoint, capabilities, true)
	publisher := mediarouter.CreatePublisher(data.Sdp)
	store.AddRouter(mediarouter)

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

	mediarouter := store.GetRouter(data.StreamID)

	if mediarouter != nil {
		mediarouter.Stop()
		store.RemoveRouter(mediarouter)
	}

	c.JSON(200, gin.H{
		"s": 10000,
		"d": map[string]string{},
	})
}

func (s *Server) unplay(c *gin.Context) {

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

	mediaRouter := store.GetRouter(data.StreamID)

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

	mediaRouter := store.GetRouter(data.StreamID)

	if mediaRouter == nil {
		c.JSON(200, gin.H{"s": 10002, "e": "can not find stream"})
		return
	}

	mediaRouter.StopSubscriber(data.SubscriberID)

	c.JSON(200, gin.H{"s": 10000, "d": map[string]string{}})
}

func (s *Server) unpullStreamFromOrigin(streamID string, subscriberID string, origin string) {

	return
}

func (s *Server) pullStreamFromOrigin(streamID string, origins []string) (*router.MediaRouter, error) {

	return nil, errors.New("can not find stream")
}

func (s *Server) startRtmp() {

}
