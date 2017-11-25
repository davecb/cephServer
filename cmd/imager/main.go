package main

import (
	"fmt"
	"html"
	ceph "imageServer/pkg/cephInterface"
	migr "imageServer/pkg/imageMigrator"
	size "imageServer/pkg/imageResizer"
	"imageServer/pkg/trace"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
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
	size.T = T

	T.Print("starting\n")
	go startWebserver()
	runLoadTest()
}

// startWebserver for all image requests
func startWebserver() {
	defer T.Begin()()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			getSizedImage(w, r)
		case "POST":
			fmt.Fprintf(w, "POST not implemented, %q", html.EscapeString(r.URL.Path))
		case "PUT":
			fmt.Fprintf(w, "PUT not implemented, %q", html.EscapeString(r.URL.Path))
		case "DELETE":
			fmt.Fprintf(w, "DELE not implemented, %q", html.EscapeString(r.URL.Path))
		default:
			fmt.Fprintf(w, "Invalid request method, %q", html.EscapeString(r.URL.Path))
			http.Error(w, "Method not allowed", 405)
		}
	})
	err := http.ListenAndServe(":8080", nil) // nolint
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// getSizedImage gets an image in a specific size
func getSizedImage(w http.ResponseWriter, r *http.Request) {
	defer T.Begin()()

	fullPath := r.URL.Path
	key, width, height, quality, grey, name, imgType, err := parseImageURL(fullPath)
	if err != nil {
		http.Error(w, "Cannot interpret url", 400)
	}
	if ceph.HaveWe(fullPath) {
		// return the file in the size requested
		io.WriteString(w, ceph.Get(fullPath)) // nolint FIXME
	} else if ceph.HaveWe(key) {
		// we have a base file which we can resize
		if width < largestWidth {
			// we can afford to do it in-line
			s := size.ResizeImage(ceph.Get(key), width, height, quality,
				grey, name, imgType)
			// return it, and save in the background
			io.WriteString(w, dummyImage(imgType)) // nolint
			go ceph.Save(fullPath, s)
		} else {
			// we background it and return a dummy
			io.WriteString(w, dummyImage(imgType)) // nolint
			go func() {
				ceph.Save(fullPath, size.ResizeImage(ceph.Get(key), width,
					height, quality, grey, name, imgType))
			}()
		}
	} else {
		// we lack a base, so background a migrate-then-resize, return a dummy
		io.WriteString(w, dummyImage(imgType)) // nolint
		go func() {
			ceph.Save(fullPath, migr.MigrateAndResizeImage(ceph.Get(key), width, height, quality,
				grey, name, imgType))
		}()

	}
}

// runLoadTest beats up the web server
func runLoadTest() {
	key := "/image/albert/100/200/85/False/albert.jpg"
	//key := "/albert.jpg"
	initial := time.Now()
	resp, err := http.Get("http://localhost:8080/" + key)
	if err != nil {
		panic(fmt.Sprintf("Got error: %v", err))
	}
	ioutil.ReadAll(resp.Body) // nolint. was body, err :=  ioutil.ReadAll(resp.Body)
	resp.Body.Close()         // nolint
	requestTime := time.Since(initial)
	reportPerformance(initial, requestTime, 0, 0, 0, 200, key)

}

// reportPerformance in standard format
func reportPerformance(initial time.Time, latency, xferTime,
	thinkTime time.Duration, length int64, rc int, key string) {

	fmt.Printf("%s %f %f %f %d %s %d GET\n",
		initial.Format("2006-01-02 15:04:05.000"),
		latency.Seconds(), xferTime.Seconds(), thinkTime.Seconds(),
		length, key, rc)
}

// return a dummy image in the appropriate type and a selected size
func dummyImage(imageType string) string {
	defer T.Begin()()
	return ""
}
