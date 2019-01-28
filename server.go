package main

import (
	"github.com/notedit/media-server-go"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"github.com/notedit/media-server-go/sdp"
)

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

var routers = map[string]*MediaRouter{}
var endpoint = mediaserver.NewEndpoint("127.0.0.1")


func publish(c *gin.Context) {

	streamID := c.Param("streamID")

	var data struct{
		Sdp string `json:"sdp"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10000, "e": err})
		return
	}

	router := NewMediaRouter(streamID, endpoint, Capabilities)

	_,answer := router.CreatePublisher(data.Sdp)

	routers[streamID] = router

	c.JSON(200,gin.H{
		"s":10000,
		"d": map[string]string{
			"sdp": answer,
		},
		"e":"",
	})
}

func unpublish(c *gin.Context) {

	streamID := c.Param("streamID")

	router,ok := routers[streamID]

	if !ok {
		c.JSON(200, gin.H{
			"s": 10000,
			"e":"stream does not exist",
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

	var data struct{
		SubscriberID string `json:"subscriberId"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10000, "e": err})
		return
	}

	router,ok := routers[streamID]

	if !ok {
		c.JSON(200, gin.H{
			"s": 10000,
			"e":"stream does not exist",
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

	var data struct{
		Sdp string `json:"sdp"`
	}

	if err := c.ShouldBind(&data); err != nil {
		c.JSON(200, gin.H{"s": 10000, "e": err})
		return
	}

	router,ok := routers[streamID]

	if !ok {
		c.JSON(200,gin.H{"s":10002, "e":"can not find stream"})
		return
	}

	subscriber, answer := router.CreateSubscriber(data.Sdp)

	c.JSON(200, gin.H{
		"s":10000,
		"d": map[string]string{
			"sdp":answer,
			"subscriberId":subscriber.GetID(),
		},
		"e":"",
	})

}


func main() {

	r := gin.Default()

	r.Use(cors.Default())
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	r.POST("/publish/:streamID", publish)
	r.POST("/unpublish/:streamID",unpublish)
	r.POST("/play/:streamID", play)
	r.POST("/unplay/:streamID",unplay)

	r.Run(":5000")
}
