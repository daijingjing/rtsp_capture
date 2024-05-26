package main

import (
	"flag"
	"fmt"
	"os"
	"rtsp_capture"
)

var (
	cfgURL  = flag.String("url", "", "RTSP URL")
	cfgFile = flag.String("out", "", "Output Image File")
)

func main() {
	flag.Parse()

	if *cfgURL == "" {
		fmt.Println("ERROR: RTSP URL is required")
		os.Exit(1)
	}
	if *cfgFile == "" {
		fmt.Println("ERROR: Output file is required")
		os.Exit(2)
	}
	err := rtsp_capture.Capture(*cfgURL, *cfgFile)
	if err != nil {
		panic(err)
	}
}
