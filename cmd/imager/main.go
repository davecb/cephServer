package main

import (

	"github.com/davecb/cephServer/pkg/imageServer"
	"github.com/davecb/cephServer/pkg/bucketServer"
	"github.com/davecb/cephServer/pkg/trace"
	
	"net/http/httputil"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	//"os"
	"time"
	"strings"
	"os"
)


// t is a debugging tool shared by the server components
var t = trace.New(os.Stderr, true) // or (nil, false)
var img = imageServer.New(t) 
var bucket = bucketServer.New(t)

const (   // FIXME Andrew's buckets
	// buckets must have leading and trailing slashes
	assets = "/assets.s3.kobo.com/"
	download = "/download.s3.kobo.com/"
	merch = "/merch.s3.kobo.com/"
	ops = "/ops.s3.kobo.com/"
)
const (
	host = "10.92.10.201:5280"    // ":5280"
)


func main() {
	defer t.Begin()()

	go runLoadTest()
	startWebserver()
}

// startWebserver for all image requests
func startWebserver() {
	defer t.Begin()()

	// handle image vs content part of prefixes
	http.HandleFunc("/images.s3.kobo.com/", imageHandler)
	http.HandleFunc(assets, func(w http.ResponseWriter, r *http.Request) {
		bucketedObjectHandler(r, w, assets)
	})
	http.HandleFunc(download, func(w http.ResponseWriter, r *http.Request) {
		bucketedObjectHandler(r, w, download)
	})
	http.HandleFunc(merch, func(w http.ResponseWriter, r *http.Request) {
		bucketedObjectHandler(r, w, merch)
	})
	http.HandleFunc(ops, func(w http.ResponseWriter, r *http.Request) {
		bucketedObjectHandler(r, w, ops)
	})
	http.HandleFunc("/", unsupportedHandler)

	// FIXME ip addr only ???
	err := http.ListenAndServe(host, nil) // nolint
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// unsupportedHandler handles bad paths
func unsupportedHandler(w http.ResponseWriter, r *http.Request) {
	defer t.Begin(r)()
	reportUnimplemented(w, "No handler for %q",	r.Method + " " + html.EscapeString(r.URL.Path))
}

// bucketedObjectHandler handles "ordinary" objects in buckets
func bucketedObjectHandler(r *http.Request, w http.ResponseWriter, bucketName string) {
	r.URL.Path = strings.TrimPrefix(r.URL.Path, bucketName)
	switch r.Method {
	case "GET":
		bucket.Get(w, r, bucketName) // FIXME, unchecked return
	case "PUT":
		// FIXME, add GET, maybe HEAD and DELETE
		reportUnimplemented(w, "PUT not implemented, %q",
			html.EscapeString(r.URL.Path))
	case "DELETE":
		reportUnimplemented(w, "DELE not implemented, %q",
			html.EscapeString(r.URL.Path))
	default:
		reportUnimplemented(w, "Method not implememted for %q",
			r.Method+" "+html.EscapeString(r.URL.Path))
	}
}


// imageHandler handles automagically-resized images
func imageHandler(w http.ResponseWriter, r *http.Request) {
	defer t.Begin(r)()

	r.URL.Path = strings.TrimPrefix(r.URL.Path,
		"/content/v1/images.s3.kobo.com/")
	switch r.Method {
	case "GET":
		img.GetSized(w, r)
	case "PUT":
		reportUnimplemented(w, "PUT not implemented, %q", html.EscapeString(r.URL.Path))
	case "DELETE":
		reportUnimplemented(w, "DELE not implemented, %q", html.EscapeString(r.URL.Path))
	default:
		reportUnimplemented(w, "Method not implememted for %q",	r.Method + " " + html.EscapeString(r.URL.Path))
	}
}

func reportUnimplemented(w http.ResponseWriter, p, q string) {
	t.Printf(p, q)
	http.Error(w, fmt.Sprintf(p, q),405)
}

// runSmokeTest checks that the server is up
func runLoadTest() {
	time.Sleep(time.Second * 2)
	key := "download.s3.kobo.com/3HK/index.html"
	//key := "download.s3.kobo.com/image/albert/100/200/85/False/albert.jpg"
	//key := "albert.jpg"
	initial := time.Now()
	resp, err := http.Get("http://" + host + "/" + key)
	if err != nil {
		panic(fmt.Sprintf("Got an error in the get: %v", err))
	}
	body, err :=  ioutil.ReadAll(resp.Body)
	requestTime := time.Since(initial)
	if err != nil {
		panic(fmt.Sprintf("Got an error in the body read: %v", err))
	}
	t.Printf("\n%s\n%s\n", responseToString(resp), bodyToString(body))
	resp.Body.Close()         // nolint
	reportPerformance(initial, requestTime, 0, 0,
		len(body),  key,resp.StatusCode, "GET")

}

// reportPerformance in standard format
func reportPerformance(initial time.Time, latency, xferTime,
	thinkTime time.Duration, length int, key string, rc int,
	op string) {

	fmt.Printf("%s %f %f %f %d %s %d %s\n",
		initial.Format("2006-01-02 15:04:05.000"),
		latency.Seconds(), xferTime.Seconds(), thinkTime.Seconds(),
		length, key, rc, op)
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
