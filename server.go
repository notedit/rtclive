package main

import (
	"encoding/json"
	"github.com/notedit/media-server-go"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/notedit/media-server-go/sdp"
	"gopkg.in/olahol/melody.v1"
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

var Capabilities = map[string]*sdp.Capability{
	"audio": &sdp.Capability{
		Codecs: []string{"opus"},
	},
	"video": &sdp.Capability{
		Codecs: []string{"vp8"},
		Rtx:    true,
		Rtcpfbs: []*sdp.RtcpFeedback{
			&sdp.RtcpFeedback{
				ID: "goog-remb",
			},
			&sdp.RtcpFeedback{
				ID: "transport-cc",
			},
			&sdp.RtcpFeedback{
				ID:     "ccm",
				Params: []string{"fir"},
			},
			&sdp.RtcpFeedback{
				ID:     "nack",
				Params: []string{"pli"},
			},
		},
		Extensions: []string{
			"urn:3gpp:video-orientation",
			"http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01",
			"http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time",
			"urn:ietf:params:rtp-hdrext:toffse",
			"urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id",
			"urn:ietf:params:rtp-hdrext:sdes:mid",
		},
	},
}

var endpoint *mediaserver.Endpoint
var config  *ConfigStruct


func publish(c *gin.Context) {

	streamID := c.Param("streamID")

	var data struct {
		Sdp string `json:"sdp"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10000, "e": err})
		return
	}

	router := NewMediaRouter(streamID, endpoint, Capabilities, true)

	_, answer := router.CreatePublisher(data.Sdp)

	routers.Add(router)

	c.JSON(200, gin.H{
		"s": 10000,
		"d": map[string]string{
			"sdp": answer,
		},
		"e": "",
	})
}

func unpublish(c *gin.Context) {

	streamID := c.Param("streamID")

	router := routers.Get(streamID)

	if router == nil {
		c.JSON(200, gin.H{
			"s": 10000,
			"e": "stream does not exist",
		})
		return
	}

	router.Stop()

	c.JSON(200, gin.H{
		"s": 10000,
		"d": map[string]string{},
		"e": "",
	})
}

func unplay(c *gin.Context) {

	streamID := c.Param("streamID")

	var data struct {
		SubscriberID string `json:"subscriberId"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10000, "e": err})
		return
	}

	router := routers.Get(streamID)

	if router == nil {
		c.JSON(200, gin.H{
			"s": 10000,
			"e": "stream does not exist",
		})
		return
	}

	router.StopSubscriber(data.SubscriberID)

	c.JSON(200, gin.H{
		"s": 10000,
		"d": map[string]string{},
		"e": "",
	})
}

func play(c *gin.Context) {

	streamID := c.Param("streamID")

	var data struct {
		Sdp string `json:"sdp"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10000, "e": err})
		return
	}

	router := routers.Get(streamID)

	if router == nil {
		c.JSON(200, gin.H{"s": 10002, "e": "can not find stream"})
		return
	}

	subscriber, answer := router.CreateSubscriber(data.Sdp)

	c.JSON(200, gin.H{
		"s": 10000,
		"d": map[string]string{
			"sdp":          answer,
			"subscriberId": subscriber.GetID(),
		},
		"e": "",
	})

}

func pull(c *gin.Context) {

	var data struct {
		StreamId string `json:"StreamId"`
		Sdp      string `json:"sdp"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10001, "e": err})
		return
	}

	router := routers.Get(data.StreamId)

	if router == nil {
		c.JSON(200, gin.H{"s": 10002, "e": "can not find stream"})
		return
	}

}

func onconnect(s *melody.Session) {
	sessions.Add(s)
}

func ondisconnect(s *melody.Session) {
	defer sessions.Remove(s)

	sessionInfo := sessions.Get(s)

	if sessionInfo.StreamID == "" {
		return
	}

	if sessionInfo.SubscriberID == "" {
		router := routers.Get(sessionInfo.StreamID)
		if router != nil {
			router.Stop()
			routers.Remove(router)
		}
	} else {
		router := routers.Get(sessionInfo.StreamID)
		if router != nil {
			router.StopSubscriber(sessionInfo.SubscriberID)
		}
	}
}

func onmessage(s *melody.Session, msg []byte) {

	var message Message

	err := json.Unmarshal(msg, &message)

	if err != nil {
		return
	}

	switch message.Cmd {
	case "publish":
		router := NewMediaRouter(message.StreamID, endpoint, Capabilities,true)
		_, answer := router.CreatePublisher(message.Sdp)
		routers.Add(router)
		sessionInfo := sessions.Get(s)
		sessionInfo.StreamID = message.StreamID
		res, _ := json.Marshal(&Response{
			Code: 0,
			Data: ResponseData{
				Sdp: answer,
			},
		})
		s.Write(res)
	case "unpublish":
		router := routers.Get(message.StreamID)
		if router == nil {
			res, _ := json.Marshal(&Response{
				Code: 1,
			})
			s.Write(res)
			return
		}
		defer routers.Remove(router)
		router.Stop()
		sessionInfo := sessions.Get(s)
		sessionInfo.StreamID = ""
		res, _ := json.Marshal(&Response{
			Code: 0,
		})
		s.Write(res)
	case "play":
		router := routers.Get(message.StreamID)
		if router == nil {
			res, _ := json.Marshal(&Response{
				Code: 1,
			})
			s.Write(res)
			return
		}
		subscriber, answer := router.CreateSubscriber(message.Sdp)
		sessionInfo := sessions.Get(s)
		sessionInfo.StreamID = message.StreamID
		sessionInfo.SubscriberID = subscriber.GetID()
		res, _ := json.Marshal(&Response{
			Code: 0,
			Data: ResponseData{
				Sdp:          answer,
				SubscriberID: subscriber.GetID(),
			},
		})
		s.Write(res)
	case "unplay":
		router := routers.Get(message.StreamID)
		if router == nil {
			res, _ := json.Marshal(&Response{
				Code: 1,
			})
			s.Write(res)
			return
		}
		router.StopSubscriber(message.SubscriberID)
		sessionInfo := sessions.Get(s)
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

func main() {

	var err error
	config,err = LoadConfig("./config.yaml")

	if err != nil {
		panic(err)
	}

	endpoint = mediaserver.NewEndpoint(config.Media.Endpoint)

	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	r.POST("/publish/:streamID", publish)
	r.POST("/unpublish/:streamID", unpublish)
	r.POST("/play/:streamID", play)
	r.POST("/unplay/:streamID", unplay)

	r.POST("/pull", pull)

	mrouter := melody.New()

	r.GET("/ws", func(c *gin.Context) {
		mrouter.HandleRequest(c.Writer, c.Request)
	})

	mrouter.HandleConnect(onconnect)

	mrouter.HandleDisconnect(ondisconnect)

	mrouter.HandleMessage(onmessage)

	r.Run(":5000")
}
