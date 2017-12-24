package imageServer

import (
	"github.com/davecb/cephServer/pkg/trace"
	"log"  

	"testing"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"os"
	"io/ioutil"
)

var tt trace.Trace
var logger *log.Logger   // FIXME this poses a problem
var imgServ *imager

const (
	nokey	  = ""
	goodkey   = "00000b30-bbc6-4315-9b0f-d003404105e3/215/60/False/beneath-him-3.jpg"
	notype    = "00000b30-bbc6-4315-9b0f-d003404105e3/215/60/False/beneath-him-3"
	noname    = "00000b30-bbc6-4315-9b0f-d003404105e3/215/60/False/.jpg"
	noname2   = "00000b30-bbc6-4315-9b0f-d003404105e3/215/60/False/jpg"
	nocolor   = "00000b30-bbc6-4315-9b0f-d003404105e3/215/60/beneath-him-3.jpg"
	noquality = "00000b30-bbc6-4315-9b0f-d003404105e3/215/False/beneath-him-3.jpg"
	nowidth   = "00000b30-bbc6-4315-9b0f-d003404105e3/60/False/beneath-him-3.jpg"
	keyonly   = "00000b30-bbc6-4315-9b0f-d003404105e3"
	nomaster  = "/no-such-file"
)

func TestGettingImages(t *testing.T) {
	var verbose = false
	if verbose {
		tt = trace.New(os.Stderr, true)
		logger = log.New(os.Stdout, "ImageServer", log.Lshortfile|log.LstdFlags)
	} else {
		tt = trace.New(ioutil.Discard, true) // os.Stderr, true) or (nil, false)
		logger = log.New(ioutil.Discard, "ImageServer", log.Lshortfile|log.LstdFlags)
	}
	tt.Begin(t)()
	imgServ = New(tt, logger)

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

	Convey("no key", t, func() {
		r, err := http.NewRequest("GET", "", nil)
		if err != nil {
			t.Fatal(err)
		}
		Convey("gets 404 no such object\n", func() {
			w := httptest.NewRecorder()
			imgServ.GetSized(w, r)
			So(w.Result().StatusCode, ShouldEqual, 404)
		})
	})

	Convey("when no type", t, func() {
		r, err := http.NewRequest("GET", notype, nil)
		if err != nil {
			t.Fatal(err)
		}
		Convey("gets 200 if it can be created\n", func() {
			w := httptest.NewRecorder()
			imgServ.GetSized(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
		})
	})

	Convey("when no name", t, func() {
		r, err := http.NewRequest("GET", noname, nil)
		if err != nil {
			t.Fatal(err)
		}
		Convey("gets 200 if it can be created\n", func() {
			w := httptest.NewRecorder()
			imgServ.GetSized(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
		})
	})
	
	Convey("when no name, second variant", t, func() {
		r, err := http.NewRequest("GET", noname2, nil)
		if err != nil {
			t.Fatal(err)
		}
		Convey("gets 200 if it can be created\n", func() {
			w := httptest.NewRecorder()
			imgServ.GetSized(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
		})
	})

	Convey("when no color", t, func() {
		r, err := http.NewRequest("GET", nocolor, nil)
		if err != nil {
			t.Fatal(err)
		}
		Convey("gets 200 if it can be created\n", func() {
			w := httptest.NewRecorder()
			imgServ.GetSized(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
		})
	})

	Convey("when no quality", t, func() {
		r, err := http.NewRequest("GET", noquality, nil)
		if err != nil {
			t.Fatal(err)
		}
		Convey("gets 200 if it can be created\n", func() {
			w := httptest.NewRecorder()
			imgServ.GetSized(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
		})
	})

	Convey("when no width", t, func() {
		r, err := http.NewRequest("GET", nowidth, nil)
		if err != nil {
			t.Fatal(err)
		}
		Convey("gets 200 if it can be created\n", func() {
			w := httptest.NewRecorder()
			imgServ.GetSized(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
		})
	})

	Convey("when just the key", t, func() {
		r, err := http.NewRequest("GET", keyonly, nil)
		if err != nil {
			t.Fatal(err)
		}
		Convey("gets 200 OK\n", func() {
			w := httptest.NewRecorder()
			imgServ.GetSized(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
		})
	})

	Convey("when no master", t,func() {
		r, err := http.NewRequest("GET", nomaster, nil)
		if err != nil {
			t.Fatal(err)
		}
		Convey("gets 404\n", func() {
			w := httptest.NewRecorder()
			imgServ.GetSized(w, r)
			So(w.Result().StatusCode, ShouldEqual, 404)
		})
	})
	

	// Not testable in the same run as the above tests.
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



