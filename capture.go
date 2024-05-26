package rtsp_capture

import (
	"errors"
	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/gortsplib/v4/pkg/format/rtph264"
	"github.com/bluenviron/mediacommon/pkg/codecs/h264"
	"github.com/pion/rtp"
	"image"
	"image/jpeg"
	"log"
	"os"
)

// This example shows how to
// 1. connect to a RTSP server and read all tracks on a path
// 2. check if there's a H264 track
// 3. decode H264 into RGBA frames
// 4. encode the frames into JPEG images and save them on disk

// This example requires the ffmpeg libraries, that can be installed in this way:
// apt install -y libavformat-dev libswscale-dev gcc pkg-config

func saveToFile(img image.Image, filename string) error {
	// create file
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	log.Println("saving", filename)

	// convert to jpeg
	return jpeg.Encode(f, img, &jpeg.Options{
		Quality: 80,
	})
}

func Capture(rtspUrl string, filename string) error {
	transport := gortsplib.TransportTCP
	c := gortsplib.Client{
		Transport: &transport,
	}

	// parse URL
	u, err := base.ParseURL(rtspUrl)
	if err != nil {
		return err
	}

	// connect to the server
	err = c.Start(u.Scheme, u.Host)
	if err != nil {
		return err
	}
	defer c.Close()

	// find available medias
	desc, _, err := c.Describe(u)
	if err != nil {
		return err
	}

	// find the H264 media and format
	var format *format.H264
	media := desc.FindFormat(&format)
	if media == nil {
		return errors.New("video track not found")
	}

	// setup RTP -> H264 decoder
	rtpDec, err := format.CreateDecoder()
	if err != nil {
		return err
	}

	// setup H264 -> raw frames decoder
	frameDec := &h264Decoder{}
	err = frameDec.initialize()
	if err != nil {
		return err
	}
	defer frameDec.close()

	// if SPS and PPS are present into the SDP, send them to the decoder
	if format.SPS != nil {
		frameDec.decode(format.SPS)
	}
	if format.PPS != nil {
		frameDec.decode(format.PPS)
	}

	// setup a single media
	_, err = c.Setup(desc.BaseURL, media, 0, 0)
	if err != nil {
		return err
	}

	iframeReceived := false

	var done = make(chan struct{})
	var doneError error = nil

	// called when a RTP packet arrives
	c.OnPacketRTP(media, format, func(pkt *rtp.Packet) {
		ok2 := true
		select {
		case _, ok2 = <-done:
			if !ok2 {
				return
			}
		default:
		}
		// extract access units from RTP packets
		au, err := rtpDec.Decode(pkt)
		if err != nil {
			if err != rtph264.ErrNonStartingPacketAndNoPrevious && err != rtph264.ErrMorePacketsNeeded {
				log.Printf("ERR: %v", err)
			}
			return
		}

		// wait for an I-frame
		if !iframeReceived {
			if !h264.IDRPresent(au) {
				log.Printf("waiting for an I-frame")
				return
			}
			iframeReceived = true
		}

		for _, nalu := range au {
			// convert NALUs into RGBA frames
			img, err := frameDec.decode(nalu)
			if err != nil {
				doneError = err
				close(done)
			}

			// wait for a frame
			if img == nil {
				continue
			}

			// convert frame to JPEG and save to file
			err = saveToFile(img, filename)
			if err != nil {
				doneError = err
				close(done)
			}

			close(done)
		}
	})

	// start playing
	_, err = c.Play(nil)
	if err != nil {
		return err
	}

	// wait
	<-done
	return doneError
}
