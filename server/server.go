package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/imroc/req"
	"github.com/notedit/RTCLive/config"
	"github.com/notedit/RTCLive/router"
	"github.com/notedit/RTCLive/store"
	mediaserver "github.com/notedit/media-server-go"
	"github.com/notedit/media-server-go/sdp"
	rtmp "github.com/notedit/rtmp-lib"
	"github.com/notedit/rtmp-lib/av"
	"github.com/olahol/melody"
)

type Message struct {
	Cmd          string `json:"cmd"`
	Sdp          string `json:"sdp,omitempty"`
	StreamID     string `json:"streamId"`
	SubscriberID string `json:"subscriberId,omitempty"`
}

type Response struct {
	Code int          `json:"code,omitempty"`
	Data ResponseData `json:"data,omitempty"`
}

type ResponseData struct {
	Sdp          string `json:"sdp,omitempty"`
	StreamID     string `json:"streamId"`
	SubscriberID string `json:"subscriberId,omitempty"`
}

type Server struct {
	melodyRouter *melody.Melody
	httpServer   *gin.Engine
	endpoint     *mediaserver.Endpoint
	rtmpServer   *rtmp.Server
	cfg          *config.Config
}

func New(cfg *config.Config) *Server {

	server := &Server{}
	server.cfg = cfg
	server.melodyRouter = melody.New()
	server.melodyRouter.Config.MaxMessageSize = 1024 * 10
	server.melodyRouter.Config.MessageBufferSize = 1024 * 10

	gin.SetMode(gin.ReleaseMode)
	httpServer := gin.Default()
	httpServer.Use(cors.Default())

	httpServer.GET("/ws", func(c *gin.Context) {
		server.melodyRouter.HandleRequest(c.Writer, c.Request)
	})

	httpServer.POST("/pull", server.pullStream)
	httpServer.POST("/unpull", server.unpullStream)

	server.httpServer = httpServer

	server.melodyRouter.HandleConnect(server.onconnect)
	server.melodyRouter.HandleDisconnect(server.ondisconnect)
	server.melodyRouter.HandleMessage(server.onmessage)

	server.endpoint = mediaserver.NewEndpoint(cfg.Media.Endpoint)

	return server
}

func (self *Server) ListenAndServe() {

	if self.cfg.Rtmp.Port > 0 {
		go self.startRtmp()
	}

	self.httpServer.Run(self.cfg.Server.Host + ":" + strconv.Itoa(self.cfg.Server.Port))
}

func (self *Server) onconnect(s *melody.Session) {
	store.AddSession(s)
}

func (self *Server) ondisconnect(s *melody.Session) {

	defer store.RemoveSession(s)

	sessionInfo := store.GetSession(s)

	if sessionInfo.StreamID == "" {
		return
	}

	if sessionInfo.SubscriberID == "" {
		router := store.GetRouter(sessionInfo.StreamID)
		if router != nil {
			router.Stop()
			store.RemoveRouter(router)
		}
	} else {
		router := store.GetRouter(sessionInfo.StreamID)
		if router != nil {
			router.StopSubscriber(sessionInfo.SubscriberID)

			if !router.IsOrgin() && len(router.GetSubscribers()) == 0 {
				self.unpullStreamFromOrigin(sessionInfo.StreamID, router.GetPublisher().GetID(), router.GetOriginUrl())
			}
		}

	}
}

func (self *Server) onmessage(s *melody.Session, msg []byte) {

	var message Message
	err := json.Unmarshal(msg, &message)
	if err != nil {
		fmt.Println("error", err)
		return
	}

	fmt.Println("message", message.Cmd, message.StreamID)

	switch message.Cmd {
	case "publish":
		capabilitys := self.cfg.Capabilities
		router := router.NewMediaRouter(message.StreamID, self.endpoint, capabilitys, true)
		publisher := router.CreateRTCPublisher(message.Sdp)
		store.AddRouter(router)
		sessionInfo := store.GetSession(s)
		sessionInfo.StreamID = message.StreamID
		res, _ := json.Marshal(&Response{
			Code: 0,
			Data: ResponseData{
				Sdp: publisher.GetAnswer(),
			},
		})
		s.Write(res)
	case "unpublish":
		router := store.GetRouter(message.StreamID)
		if router == nil {
			res, _ := json.Marshal(&Response{
				Code: 1,
			})
			s.Write(res)
			return
		}
		defer store.RemoveRouter(router)
		router.Stop()
		sessionInfo := store.GetSession(s)
		sessionInfo.StreamID = ""
		res, _ := json.Marshal(&Response{
			Code: 0,
		})
		s.Write(res)
	case "play":
		router := store.GetRouter(message.StreamID)

		if router == nil {
			if self.cfg.Cluster.Origins != nil {
				var err error
				router, err = self.pullStreamFromOrigin(message.StreamID, self.cfg.Cluster.Origins)
				if err != nil {
					fmt.Println(err)
					panic(err)
				}
			}
		}

		if router == nil {
			res, _ := json.Marshal(&Response{
				Code: 1,
			})
			s.Write(res)
			return
		}
		subscriber, answer := router.CreateSubscriber(message.Sdp)
		sessionInfo := store.GetSession(s)
		sessionInfo.StreamID = message.StreamID
		sessionInfo.SubscriberID = subscriber.GetID()
		res, _ := json.Marshal(&Response{
			Code: 0,
			Data: ResponseData{
				Sdp:          answer,
				SubscriberID: subscriber.GetID(),
			},
		})
		fmt.Println(res)
		s.Write(res)
	case "unplay":
		router := store.GetRouter(message.StreamID)
		if router == nil {
			res, _ := json.Marshal(&Response{
				Code: 1,
			})
			s.Write(res)
			return
		}
		router.StopSubscriber(message.SubscriberID)
		sessionInfo := store.GetSession(s)
		sessionInfo.StreamID = ""
		sessionInfo.SubscriberID = ""
		res, _ := json.Marshal(&Response{
			Code: 0,
		})
		s.Write(res)

	default:
		return
	}
}

func (self *Server) pullStream(c *gin.Context) {

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

	subscriber, answer := mediaRouter.CreateSubscriber(data.Sdp)

	fmt.Println("answer", answer)

	c.JSON(200, gin.H{"s": 10000, "d": map[string]string{
		"sdp":          answer,
		"subscriberId": subscriber.GetID(),
	}})
}

func (self *Server) unpullStream(c *gin.Context) {

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

func (self *Server) unpullStreamFromOrigin(streamID string, subscriberID string, origin string) {

	var requestUrl string
	if !strings.HasPrefix(origin, "http") {
		requestUrl = "http://" + origin + "/unpull"
	} else {
		requestUrl = origin + "/unpull"
	}

	_, err := req.Post(requestUrl, req.BodyJSON(map[string]string{
		"streamId":     streamID,
		"subscriberId": subscriberID,
	}))

	if err != nil {
		fmt.Println(err)
	}

	return
}

func (self *Server) pullStreamFromOrigin(streamID string, origins []string) (*router.MediaRouter, error) {

	offer := self.endpoint.CreateOffer(self.cfg.Capabilities["video"], self.cfg.Capabilities["audio"])

	for _, origin := range origins {
		var requestUrl string
		if !strings.HasPrefix(origin, "http") {
			requestUrl = "http://" + origin + "/pull"
		} else {
			requestUrl = origin + "/pull"
		}

		res, err := req.Post(requestUrl, req.BodyJSON(map[string]string{
			"streamId": streamID,
			"sdp":      offer.String(),
		}))

		if err != nil {
			fmt.Println(err)
			continue
		}

		var ret struct {
			Status int               `json:"s"`
			Data   map[string]string `json:"d"`
		}

		err = res.ToJSON(&ret)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if ret.Status > 10000 {
			continue
		}

		answerStr := ret.Data["sdp"]
		if len(answerStr) == 0 {
			continue
		}

		answer, err := sdp.Parse(answerStr)

		if err != nil {
			fmt.Println("parse error", err)
			panic(err)
		}

		if answer.GetFirstStream() == nil {
			panic("can not get stream info")
		}

		capabilitys := self.cfg.Capabilities

		mediaRouter := router.NewMediaRouter(streamID, self.endpoint, capabilitys, false)

		mediaRouter.CreateRelayPublisher(offer.String(), answer.String())

		mediaRouter.SetOriginUrl(origin)

		return mediaRouter, nil
	}

	return nil, errors.New("can not find stream")
}

func (self *Server) startRtmp() {

	self.rtmpServer = &rtmp.Server{}

	self.rtmpServer.HandlePublish = func(conn *rtmp.Conn) {

		streaminfo := strings.Split(conn.URL.Path, "/")

		if len(streaminfo) <= 2 {
			fmt.Println("rtmp url does not match, rtmp url should like rtmp://host:/appname/streamname")
			conn.Close()
			return
		}

		streamName := streaminfo[len(streaminfo)-1]

		capabilitys := self.cfg.Capabilities

		mediaRouter := router.NewMediaRouter(streamName, self.endpoint, capabilitys, true)

		publisher := mediaRouter.CreateRTMPPublisher(streamName)
		store.AddRouter(mediaRouter)

		var streams []av.CodecData
		var err error

		if streams, err = conn.Streams(); err != nil {
			fmt.Println(err)
			return
		}

		if err = publisher.WriteHeader(streams); err != nil {
			fmt.Println(err)
			return
		}

		for {
			packet, err := conn.ReadPacket()
			if err != nil {
				fmt.Println(err)
				break
			}
			publisher.WritePacket(packet)
		}

		store.RemoveRouter(mediaRouter)
		mediaRouter.Stop()

	}

	self.rtmpServer.ListenAndServe()
}
