package main

import (
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"strings"

	ceph "imageServer/pkg/cephInterface"
	migr "imageServer/pkg/imageMigrator"
	resize "imageServer/pkg/imageResizer"
	"imageServer/pkg/trace"
)

const largestWidth = 100

// T is a debugging tool shared by the server components
var T trace.Trace

// courtesy of  http://tleyden.github.io/blog/2016/11/21/tuning-the-go-http-client-library-for-load-testing/
func main() {
	T = trace.New(os.Stderr, true)
	defer T.Begin()()
	ceph.T = T
	migr.T = T
	resize.T = T

	go runLoadTest()
	startWebserver()

}

// startWebserver for all image requests
func startWebserver() {
	defer T.Begin()()

	// fixme, should be image
	http.HandleFunc("/content/v1/images.s3.kobo.com/", func(w http.ResponseWriter, r *http.Request) {
		defer T.Begin(r)()

		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/content/v1/images.s3.kobo.com/")
		switch r.Method {
		case "GET":
			getSizedImage(w, r)
		case "POST":
			fmt.Fprintf(w,"POST not implemented, %q", html.EscapeString(r.URL.Path))
		case "PUT":
			fmt.Fprintf(w, "PUT not implemented, %q", html.EscapeString(r.URL.Path))
		case "DELETE":
			fmt.Fprintf(w, "DELE not implemented, %q", html.EscapeString(r.URL.Path))
		default:
			fmt.Fprintf(w, "Invalid request method, %q", html.EscapeString(r.URL.Path))
			http.Error(w, "Method not allowed", 405)
		}
	})
	// handler for a shorter prefix,content, etc, etc

	err := http.ListenAndServe(":8081", nil) // nolint
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// getSizedImage gets an image in a specific resize
func getSizedImage(w http.ResponseWriter, r *http.Request) {
	defer T.Begin(r.URL.Path)()

	fullPath := r.URL.Path
	key, width, height, quality, grey, name, imgType, err := parseImageURL(fullPath)
	if err != nil {
		http.Error(w, "Cannot interpret url " + fullPath , 400)
	}

	bytes, err := ceph.Get(fullPath)
	if err == nil {
		// return the file in the resize requested
		io.WriteString(w, bytes) // nolint
		return
	}

	bytes, err = ceph.Get(key)
	if err == nil {
		// we have a base file which we can resize
		if width < largestWidth {
			// we can afford to do it in-line
			s := resize.Image(bytes, key, width, height,
				quality, grey, name, imgType)
			// return it, and save in the background
			io.WriteString(w, dummyImage(imgType)) // nolint
			go ceph.Save(s, fullPath)
		} else {
			// we background it and return a dummy FIXME or the original
			io.WriteString(w, dummyImage(imgType)) // nolint
			go func() {
				ceph.Save(resize.Image(bytes, key,
					width, height, quality, grey, name, imgType), fullPath)
			}()
		}
		return
	}

	// we lack a base, so background a migrate-then-resize, return a dummy
	io.WriteString(w, dummyImage(imgType)) // nolint
	go func() {
		ceph.Save(migr.MigrateAndResizeImage(bytes,
			key, width, height, quality, grey, name, imgType),fullPath)
	}()
}

// runLoadTest beats on the web server
func runLoadTest() {
	time.Sleep(time.Second * 2)
	key := "/content/v1/images.s3.kobo.com/image/albert/100/200/85/False/albert.jpg"
	//key := "/albert.jpg"
	initial := time.Now()
	resp, err := http.Get("http://localhost:8081/" + key)
	if err != nil {
		panic(fmt.Sprintf("Got an error in the get: %v", err))
	}
	body, err :=  ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Got an error in the body read: %v", err))
	}
	resp.Body.Close()         // nolint
	requestTime := time.Since(initial)
	reportPerformance(initial, requestTime, 0, 0, len(body), 200, key)

}

// reportPerformance in standard format
func reportPerformance(initial time.Time, latency, xferTime,
	thinkTime time.Duration, length int, rc int, key string) {

	fmt.Printf("%s %f %f %f %d %s %d GET\n",
		initial.Format("2006-01-02 15:04:05.000"),
		latency.Seconds(), xferTime.Seconds(), thinkTime.Seconds(),
		length, key, rc)
}

// return a dummy image in the appropriate type and a selected resize
func dummyImage(imageType string) string {
	defer T.Begin()()
	return "dummy image"
}
