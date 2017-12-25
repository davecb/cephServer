package main

import (

	"github.com/davecb/cephServer/pkg/imageServer"
	"github.com/davecb/cephServer/pkg/objectServer"
	"github.com/davecb/trace"
	
	"net/http/httputil"
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"time"
	"strings"
	"os"
)


// t is a debugging tool shared by the server components
var t = trace.New(os.Stderr, true) // or (nil, false)
// logger goes to stdout, as do timing records for each access
var logger = log.New(os.Stdout, "", log.Ldate | log.Ltime | log.Lshortfile)
var img = imageServer.New(t, logger)
var bucket = objectServer.New(t, logger)

const (   // FIXME Andrew's bucket names
	// buckets must have leading and trailing slashes
	images = "/images.s3.kobo.com/"
	assets = "/assets.s3.kobo.com/"
	download = "/download.s3.kobo.com/"
	merch = "/merch.s3.kobo.com/"
	ops = "/ops.s3.kobo.com/"
)
const (
	//host = "10.92.10.201:5280"
	host = ":5280"
)


// main starts the web server, and also a smoke test for it
func main() {
	defer t.Begin()()

	go runSmokeTest()
	startWebserver()
}

// startWebserver for all object requests
func startWebserver() {
	defer t.Begin()()

	// handle image vs content part of prefixes
	http.HandleFunc(images, imageHandler)
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
		bucket.Get(w, r, bucketName)
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

	r.URL.Path = strings.TrimPrefix(r.URL.Path,	images)
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

// runSmokeTest checks that the server is up, panics if not
func runSmokeTest() {
	time.Sleep(time.Second * 2)
	//key := "download.s3.kobo.com/3HK/index.html"  // valid
	//key := "download.s3.kobo.com/image/albert/100/200/85/False/albert.jpg"  // TBA
	//key := "albert.jpg" // no bucket, fail 404
	//key := "download.s3.kobo.com/absent-file.jpg"  // Invalid, 404
	//key := "images.s3.kobo.com/image/albert/100/200/85/False/albert"
	key := "images.s3.kobo.com/00000b30-bbc6-4315-9b0f-d003404105e3"

	resp, err := http.Get("http://" + host + "/" + key)
	if err != nil {
		panic(fmt.Sprintf("Got an error in the get: %v", err))
	}
	body, err :=  ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("Got an error in the body read: %v", err))
	}
	t.Printf("\n%s\n%s\n", responseToString(resp), bodyToString(body))
	resp.Body.Close()         // nolint
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
