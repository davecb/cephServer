package imageServer

import (
	"github.com/davecb/cephServer/pkg/trace"
	"log"  

	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	//"io/ioutil" // for ioutil.Discard
	"os"
)

var tt = trace.New(os.Stderr, true) // os.Stderr, true) or (nil, false)
var logger = log.New(os.Stdout, "ImageServer", log.Lshortfile|log.LstdFlags)
var imgServ *imager
const (
	goodpath = "/3HK/index.html"
	badpath = "/no-such-file"

	goodkey = "/image/albert/100/200/85/False/albert.jpg"
	notype = "/image/albert/100/200/85/False/albert"
	noname = "/image/albert/100/200/85/False/.jpg"
	noname2 = "/image/albert/100/200/85/False/jpg"
	nocolor = "/image/albert/100/200/85/albert.jpg"
	noquality = "/image/albert/100/200/False/albert.jpg"
	noheight = "/image/albert/100/85/False/albert.jpg"
	nowidth = "/image/albert/200/85/False/albert.jpg"
	nokey = "//100/200/85/False/albert.jpg"
)

func TestGettingImages(t *testing.T) {
	tt.Begin(t)()
	imgServ = New(tt, logger)

	//Convey("When all good\n", t, func() {
	//	rGood, err := http.NewRequest("GET", goodpath, nil)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	Convey("gets a 200 OK\n", func() {
	//		w := httptest.NewRecorder()
	//		imgServ.GetSized(w, rGood,)
	//		So(w.Result().StatusCode, ShouldEqual, 200)
	//	})
	//})

	//Convey("When path bad\n", t, func() {
	//	r, err := http.NewRequest("GET", badpath, nil)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	Convey("gets 200 and the dummy image\n", func() {
	//		w := httptest.NewRecorder()
	//		imgServ.GetSized(w, r)
	//		So(w.Result().StatusCode, ShouldEqual, 200)
	//		So(w.Body.String(), ShouldEqual, "dummy image")
	//	})
	//})
	

	Convey("good long image key", t, func() {
		r, err := http.NewRequest("GET", goodkey, nil)
		if err != nil {
			t.Fatal(err)
		}
		Convey("gets 200 OK\n", func() {
			w := httptest.NewRecorder()
			imgServ.GetSized(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
		})
	})
	//
	//Convey("when no type", t, func() {
	//	r, err := http.NewRequest("GET", notype, nil)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	Convey("gets 404, too\n", func() {
	//		w := httptest.NewRecorder()
	//		imgServ.GetSized(w, r)
	//		So(w.Result().StatusCode, ShouldEqual, 404)
	//	})
	//
	//})
	//Convey("when no name", t, func() {
	//	r, err := http.NewRequest("GET", noname, nil)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	Convey("gets 404, too\n", func() {
	//		w := httptest.NewRecorder()
	//		imgServ.GetSized(w, r)
	//		So(w.Result().StatusCode, ShouldEqual, 404)
	//	})
	//
	//})
	//Convey("when no name, second variant", t, func() {
	//	r, err := http.NewRequest("GET", noname2, nil)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	Convey("gets 404, too\n", func() {
	//		w := httptest.NewRecorder()
	//		imgServ.GetSized(w, r)
	//		So(w.Result().StatusCode, ShouldEqual, 404)
	//	})
	//
	//})
	//Convey("when no color", t, func() {
	//	r, err := http.NewRequest("GET", nocolor, nil)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	Convey("gets 404, too\n", func() {
	//		w := httptest.NewRecorder()
	//		imgServ.GetSized(w, r)
	//		So(w.Result().StatusCode, ShouldEqual, 404)
	//	})
	//
	//})
	//Convey("when no quality", t, func() {
	//	r, err := http.NewRequest("GET", noquality, nil)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	Convey("gets 404, too\n", func() {
	//		w := httptest.NewRecorder()
	//		imgServ.GetSized(w, r)
	//		So(w.Result().StatusCode, ShouldEqual, 404)
	//	})
	//
	//})
	//Convey("when no width", t, func() {
	//	r, err := http.NewRequest("GET", nowidth, nil)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	Convey("gets 404, too\n", func() {
	//		w := httptest.NewRecorder()
	//		imgServ.GetSized(w, r)
	//		So(w.Result().StatusCode, ShouldEqual, 404)
	//	})
	//
	//})
	//Convey("when no height", t, func() {
	//	r, err := http.NewRequest("GET", noheight, nil)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	Convey("gets 404, too\n", func() {
	//		w := httptest.NewRecorder()
	//		imgServ.GetSized(w, r)
	//		So(w.Result().StatusCode, ShouldEqual, 404)
	//	})
	//
	//})
	//Convey("when no key", t, func() {
	//	r, err := http.NewRequest("GET", nokey, nil)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	Convey("gets 404, too\n", func() {
	//		w := httptest.NewRecorder()
	//		imgServ.GetSized(w, r)
	//		So(w.Result().StatusCode, ShouldEqual, 404)
	//	})
	//
	//})


	//Convey("when no precreated", t,func() {
	//	// So(w.Result().StatusCode, ShouldEqual, 202)
	//})
	//
	//Convey("when no master", t,func() {
	//	// So(w.Result().StatusCode, ShouldEqual, 202)
	//})
	//
	//Convey("when no mogile", t, func() {
	//	// this may be unreachable/untestable...
	//	// So(w.Result().StatusCode, ShouldEqual, 404)
	//})



	//Convey("When server down\n", t, func() {
	//	rGood, err := http.NewRequest("GET", goodpath, nil)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	Convey("gets a timeout\n", func() {
	//		w := httptest.NewRecorder()
	//		imgServ.Get(w, rGood, goodbucket)
	//		rc := w.Result().StatusCode
	//		So(rc, ShouldEqual, 500)
	//	})
	//
	//})

}



