package transcoder

const audio2rtp = "appsrc do-timestamp=true is-live=true name=appsrc ! decodebin ! audioconvert ! audioresample ! opusenc ! rtpopuspay timestamp-offset=0 pt=%d ! appsink name=appsink"
const video2rtp = "appsrc do-timestamp=true is-live=true name=appsrc ! h264parse ! rtph264pay timestamp-offset=0 config-interval=-1 pt=%d ! appsink name=appsink"
