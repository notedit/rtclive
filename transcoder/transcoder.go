package transcoder

import (
	"bytes"

	"github.com/notedit/rtmp-lib/av"

	"github.com/notedit/gstreamer-go"
	"github.com/notedit/rtmp-lib/aac"
	"github.com/notedit/rtmp-lib/h264"
)

type Transcoder interface {
}

type RTMPTranscoder struct {
	stream        string
	app           string
	pipeline      *gstreamer.Pipeline
	videoPipeline *gstreamer.Pipeline
	adtsheader    []byte

	videoWriteBuffer bytes.Buffer
	audioWriteBuffer bytes.Buffer

	videoCodecData h264.CodecData
	audioCodecData aac.CodecData
}

func (t *RTMPTranscoder) WritePacket(packet av.Packet) error {
	return nil
}
