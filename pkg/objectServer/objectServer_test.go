package objectServer

import (
	"github.com/davecb/cephServer/pkg/trace"
	"log"  

	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"io/ioutil"
)
var tt = trace.New(nil, false) // or (nil, false)
var logger = log.New(ioutil.Discard, "ImageServer", log.Lshortfile|log.LstdFlags)
var objServ *Object
const (
	goodpath = "/3HK/index.html"
	badpath = "/no-such-file"
	goodbucket = "download.s3.kobo.com"
	badbucket = "no-such-bucket.s3.kobo.com"
)

func TestGettingObjects(t *testing.T) {
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
		Convey("gets 404, maybe\n", func() {
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



