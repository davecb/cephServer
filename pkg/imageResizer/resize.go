package imageResizer

import (
	"imageServer/pkg/trace"

	"github.com/nfnt/resize"
	"image"
	"os"
	"image/jpeg"
	"log"
	"bytes"
	"image/png"
	"time"
	"fmt"
)

// T is a debugging tool shared by the server components
var T trace.Trace 
var sample  image.Image

// init sets up the test with a single image
func init() {
	file, err := os.Open("01.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close() // nolint
	// decode jpeg into image.Image
	img, err := jpeg.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	sample = img
}

// ResizeImage strictly resizes an image.
// FIXME we pass in contents which we don't use
func ResizeImage(contents string, width, height, quality uint, grayScale bool, name, imgType string) string {
	defer T.Begin("<contents>", width, height, quality, grayScale, name, imgType)()

	buf := new(bytes.Buffer)

	initial := time.Now()
	m := resize.Resize(width, height, sample, resize.NearestNeighbor)
	switch {
	case imgType == "jpg":
		opt := jpeg.Options{Quality: int(quality)}
		err := jpeg.Encode(buf, m, &opt)
		if err != nil {
			log.Fatalf("jpg write failure %v\n", err)
		}
	case imgType == "png":
		err := png.Encode(buf, m)
		if err != nil {
			log.Fatalf("png write failure %v\n", err)
		}
		//ico
		// jpg
		// pdf
		// png

	default:
		log.Fatal("not a jpg") // FIXME
	}
	resizeTime := time.Since(initial)
	reportPerformance(initial, resizeTime, 0,0, 0, 200, "-")
	return buf.String()
}

// reportPerformance in standard format
func reportPerformance(initial time.Time, latency, xferTime,
		thinkTime time.Duration, length int64, rc int, key string) {

	fmt.Printf("%s %f %f %f %d %s %d RESIZE\n",
	initial.Format("2006-01-02 15:04:05.000"),
	latency.Seconds(), xferTime.Seconds(), thinkTime.Seconds(),
	length, key, rc)
}

