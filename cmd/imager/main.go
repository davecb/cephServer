package main

import (
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
	"strings"

	ceph "github.com/davecb/cephServer/pkg/cephInterface"
	migr "github.com/davecb/cephServer/pkg/imageMigrator"
	resize "github.com/davecb/cephServer/pkg/imageResizer"
	bucket "github.com/davecb/cephServer/pkg/bucketServers"
	"github.com/davecb/cephServer/pkg/trace"
	"net/http/httputil"
)


// T is a debugging tool shared by the server components
var T trace.Trace

func main() {
	T = trace.New(os.Stderr, true)
	defer T.Begin()()
	ceph.T = T
	migr.T = T
	resize.T = T
	bucket.T = T

	go runLoadTest()
	startWebserver()
}

// startWebserver for all image requests
func startWebserver() {
	defer T.Begin()()

	// handle image vs content part of prefixes
	http.HandleFunc("/images/v1/images.s3.kobo.com/", imageHandler)       // FIXME Andrew's notation
	http.HandleFunc("/content/v1/download.s3.kobo.com/", objectHandler)
	http.HandleFunc("/", unsupportedHandler)

	err := http.ListenAndServe(":8081", nil) // nolint
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// unsupportedHandler handles bad paths
func unsupportedHandler(w http.ResponseWriter, r *http.Request) {
	defer T.Begin(r)()

	http.Error(w, r.Method + " " + html.EscapeString(r.URL.Path) + " not found", 404)
}

// objectHandler handles "ordinary" objects
func objectHandler(w http.ResponseWriter, r *http.Request) {
	defer T.Begin(r)()

	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/content/v1/images.s3.kobo.com/")
	switch r.Method {
	case "GET":
		bucket.Get(w, r)
	case "PUT":
		fmt.Fprintf(w, "PUT not implemented, %q", html.EscapeString(r.URL.Path))
	case "DELETE":
		fmt.Fprintf(w, "DELE not implemented, %q", html.EscapeString(r.URL.Path))
	default:
		http.Error(w, fmt.Sprintf("Method %q not implememted for %q",
			r.Method, html.EscapeString(r.URL.Path)), 405)
	}
}


// imageHandler handles automagically-resized images
func imageHandler(w http.ResponseWriter, r *http.Request) {
	defer T.Begin(r)()

	r.URL.Path = strings.TrimPrefix(r.URL.Path, "/content/v1/images.s3.kobo.com/")
	switch r.Method {
	case "GET":
		bucket.GetSizedImage(w, r)
	case "PUT":
		fmt.Fprintf(w, "PUT not implemented, %q", html.EscapeString(r.URL.Path))
	case "DELETE":
		fmt.Fprintf(w, "DELE not implemented, %q", html.EscapeString(r.URL.Path))
	default:
		fmt.Fprintf(w, fmt.Sprintf("Method %q not implememted for %q",
			r.Method, html.EscapeString(r.URL.Path)), 405)
	}
}

// runLoadTest beats on the web server
func runLoadTest() {
	time.Sleep(time.Second * 2)
	key := "/content/v1/download.s3.kobo.com/image/albert/100/200/85/False/albert.jpg"
	//key := "/albert.jpg"
	initial := time.Now()
	resp, err := http.Get("http://localhost:8081/" + key)
	if err != nil {
		panic(fmt.Sprintf("Got an error in the get: %v", err))
	}
	body, err :=  ioutil.ReadAll(resp.Body)
	requestTime := time.Since(initial)
	if err != nil {
		panic(fmt.Sprintf("Got an error in the body read: %v", err))
	}
	T.Printf("\n%s\n%s\n", responseToString(resp), bodyToString(body))
	resp.Body.Close()         // nolint
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

// requestToString provides extra information about an http request if it can
func requestToString(req *http.Request) string {
	var dump []byte
	var err error

	if req == nil {
		return "Request: <nil>\n"
	}
	dump, err = httputil.DumpRequestOut(req, true)
	if err != nil {
		return fmt.Sprintf("fatal error dumping http request, %v\n", err)
	}
	return fmt.Sprintf("Request: \n%s", dump)
}

// responseToString provides extra information about an http response
func responseToString(resp *http.Response) string {
	if resp == nil {
		return "Response: <nil>\n"
	}
	s := requestToString(resp.Request)
	contents, err := httputil.DumpResponse(resp, false)
	if err != nil {
		return fmt.Sprintf("error dumping http response, %v\n", err)
	}
	s += "Response information:\n"
	s += fmt.Sprintf("    Length: %d\n", resp.ContentLength)
	s += fmt.Sprintf("    Status code: %d\n", resp.StatusCode)
	s += fmt.Sprintf("Response contents: \n%s", string(contents))
	return s
}

// bodyToString
func bodyToString(body []byte) string {
	if body == nil {
		return "Body: <nil>\n"
	}
	return fmt.Sprintf("Body:\n %s\n", body)
}
