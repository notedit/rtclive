package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"runtime"
	"strconv"
	"strings"

	cconfig "github.com/notedit/RTCLive/config"
	"github.com/notedit/RTCLive/rtmpstreamer"

	"github.com/imroc/req"
	"github.com/notedit/media-server-go"
	"github.com/notedit/media-server-go/sdp"
	"gopkg.in/olahol/melody.v1"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nareix/joy4/format/rtmp"
)

func init() {
	runtime.GOMAXPROCS(1)
}

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

var endpoint *mediaserver.Endpoint
var config *cconfig.Config

func pull(c *gin.Context) {

	var data struct {
		StreamID string `json:"streamId"`
		Sdp      string `json:"sdp"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10001, "e": err})
		return
	}

	router := routers.Get(data.StreamID)

	if router == nil {
		c.JSON(200, gin.H{"s": 10002, "e": "can not find stream"})
		return
	}

	subscriber, answer := router.CreateSubscriber(data.Sdp)

	fmt.Println("answer", answer)

	c.JSON(200, gin.H{"s": 10000, "d": map[string]string{
		"sdp":          answer,
		"subscriberId": subscriber.GetID(),
	}})
}

func unpull(c *gin.Context) {

	var data struct {
		StreamID     string `json:"streamId"`
		SubscriberID string `json:"subscriberId"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10001, "e": err})
		return
	}

	router := routers.Get(data.StreamID)

	if router == nil {
		c.JSON(200, gin.H{"s": 10002, "e": "can not find stream"})
		return
	}

	router.StopSubscriber(data.SubscriberID)

	c.JSON(200, gin.H{"s": 10000, "d": map[string]string{}})
}

func onconnect(s *melody.Session) {
	sessions.Add(s)
	fmt.Println("onconnect")
}

func ondisconnect(s *melody.Session) {

	fmt.Println("ondisconnect")

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

			if !router.IsOrgin() && len(router.subscribers) == 0 {
				unpullStream(sessionInfo.StreamID, router.GetPublisher().GetID(), router.GetOriginUrl())
			}
		}

	}
}

func onmessage(s *melody.Session, msg []byte) {

	var message Message
	err := json.Unmarshal(msg, &message)
	if err != nil {
		fmt.Println("error", err)
		return
	}

	fmt.Println("message", message.Cmd, message.StreamID)

	switch message.Cmd {
	case "publish":
		capabilitys := map[string]*sdp.Capability{
			"video": config.VideoCapability,
			"audio": config.AudioCapability,
		}
		router := NewMediaRouter(message.StreamID, endpoint, capabilitys, true)
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
			if config.Cluster.Origins != nil {
				var err error
				router, err = pullStream(message.StreamID, config.Cluster.Origins)
				if err != nil {
					fmt.Println(err)
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

func pullStream(streamID string, origins []string) (*MediaRouter, error) {

	offer := endpoint.CreateOffer(config.VideoCapability, config.AudioCapability)

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

		fmt.Println(answer)
		fmt.Println(answer.GetAudioMedia())
		fmt.Println(answer.GetVideoMedia())

		if answer.GetFirstStream() == nil {
			panic("can not get stream info")
		}
		fmt.Println(answer.GetFirstStream())

		transport := endpoint.CreateTransport(answer, offer, true)

		transport.SetLocalProperties(offer.GetAudioMedia(), offer.GetVideoMedia())
		transport.SetRemoteProperties(answer.GetAudioMedia(), answer.GetVideoMedia())

		streamInfo := answer.GetFirstStream()

		incoming := transport.CreateIncomingStream(streamInfo)

		capabilitys := map[string]*sdp.Capability{
			"video": config.VideoCapability,
			"audio": config.AudioCapability,
		}

		router := NewMediaRouter(streamID, endpoint, capabilitys, false)

		publisher := NewPublisher(incoming, transport)

		router.SetPublisher(publisher)

		router.SetOriginUrl(origin)

		return router, nil
	}

	return nil, errors.New("can not find stream")
}

func unpullStream(streamID string, subscriberID string, origin string) {

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

func startRtmp() {

	server := &rtmp.Server{}

	server.HandlePlay = func(conn *rtmp.Conn) {}

	server.HandlePublish = func(conn *rtmp.Conn) {

		//streamId := conn.URL.Path
		streaminfo := strings.Split(conn.URL.Path, "/")
		if len(streaminfo) != 2 {
			panic("rtmpurl invalide")
		}

		streamName := streaminfo[1]

		rtmpStreamer := rtmpstreamer.NewRtmpStreamer(streamName, config.AudioCapability, config.VideoCapability)

		avutil.CopyFile(rtmpStreamer, conn)


		writeheader := make(chan bool)
		done := make(chan bool)

		go func() {
			var streams []av.CodecData
			var err error
			if streams,err = conn.Streams(); err != nil {
				fmt.Println(err)
				done <- true
				return
			}

			if err = rtmpStreamer.WriteHeader(streams); err != nil {
				done <- true
				return
			} else {
				writeheader <- true
			}
			if err = avutil.CopyPackets(rtmpStreamer, conn); err != nil {
				fmt.Println(err)
				done <- true
				return
			}

		}()


		for {
			select {
			case <- done:
				break
			case <- writeheader:
				// todo 

			}
		}

	}
}

func main() {

	var err error
	config, err = cconfig.LoadConfig("./config.yaml")

	if err != nil {
		panic(err)
	}

	endpoint = mediaserver.NewEndpoint(config.Media.Endpoint)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(cors.Default())

	r.POST("/pull", pull)
	r.POST("/unpull", unpull)

	mrouter := melody.New()
	mrouter.Config.MaxMessageSize = 1024 * 10
	mrouter.Config.MessageBufferSize = 1024 * 5

	r.GET("/ws", func(c *gin.Context) {
		mrouter.HandleRequest(c.Writer, c.Request)
	})

	mrouter.HandleConnect(onconnect)

	mrouter.HandleDisconnect(ondisconnect)

	mrouter.HandleMessage(onmessage)

	r.Run(config.Server.Host + ":" + strconv.Itoa(config.Server.Port))
}
