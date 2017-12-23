package objectServer

import (
	"github.com/davecb/cephServer/pkg/trace"
	"log"  

	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"io/ioutil"
	"os"
)


var objServ *Object
var tt trace.Trace
var logger *log.Logger   // FIXME this poses a problem


const (
	goodpath = "/3HK/index.html"
	badpath = "/no-such-file"
	goodbucket = "download.s3.kobo.com"
	badbucket = "no-such-bucket"
)

func TestGettingObjects(t *testing.T) {
	var verbose = false
	if verbose {
		tt = trace.New(os.Stderr, true)
		logger = log.New(os.Stdout, "ImageServer", log.Lshortfile|log.LstdFlags)
	} else {
		tt = trace.New(ioutil.Discard, true) // os.Stderr, true) or (nil, false)
		logger = log.New(ioutil.Discard, "ImageServer", log.Lshortfile|log.LstdFlags)
	}
	tt.Begin(t)()
	objServ = New(tt, logger)

	Convey("When all good\n", t, func() {
		rGood, err := http.NewRequest("GET", goodpath, nil)
		if err != nil {
			t.Fatal(err)
		}
		Convey("gets a 200 OK\n", func() {
			w := httptest.NewRecorder()
			objServ.Get(w, rGood, goodbucket)
			So(w.Result().StatusCode, ShouldEqual, 200)
		})
	})

	Convey("When path bad\n", t, func() {
		rBad, err := http.NewRequest("GET", badpath, nil)
		if err != nil {
			t.Fatal(err)
		}
		Convey("gets a 404 no such file \n", func() {
			w := httptest.NewRecorder()
			objServ.Get(w, rBad, goodbucket)
			So(w.Result().StatusCode, ShouldEqual, 404)
		})
	})

	Convey("When bucket bad\n", t, func() {
		rBad, err := http.NewRequest("GET", goodpath, nil)
		if err != nil {
			t.Fatal(err)
		}
		Convey("gets 404, too\n", func() {
			w := httptest.NewRecorder()
			objServ.Get(w, rBad, badbucket)
			So(w.Result().StatusCode, ShouldEqual, 404)
		})


	})

	//Convey("When server down\n", t, func() {
	//	rGood, err := http.NewRequest("GET", goodpath, nil)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	Convey("gets a timeout\n", func() {
	//		w := httptest.NewRecorder()
	//		objServ.Get(w, rGood, goodbucket)
	//		rc := w.Result().StatusCode
	//		So(rc, ShouldEqual, 500)
	//	})
	//
	//})

}



