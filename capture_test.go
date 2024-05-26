package rtsp_capture

import (
	"strconv"
	"testing"
	"time"
)

func TestCapture(t *testing.T) {
	fname := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10) + ".jpg"

	err := Capture("rtsp://admin:1234qwer@192.168.6.171/h264/ch1/main/av_stream", fname)
	if err != nil {
		t.Log("error:", err)
	}
}
