package objectServer


import (
	"net/http"
	"github.com/davecb/cephServer/pkg/trace"
	"github.com/davecb/cephServer/pkg/cephInterface"
	"log"
)

var ceph *cephInterface.S3Proto   // maybe move

// Object is a storage bucket
type Object struct {
	logger *log.Logger
}
var t trace.Trace

// New creates a object server
func New(x trace.Trace, y *log.Logger) *Object {
	ceph = cephInterface.New(x, y)
	t = x
	return &Object{ y }
}

// Get an object from a specific bucket. Errors are written to w
func (o Object) Get(w http.ResponseWriter, r *http.Request, bucket string)  { // nolint
	var head map[string]string
	defer t.Begin(r.URL.Path)()

	t.Printf("got a request for %s\n", r.URL.Path)
	data, head, rc, err := ceph.Get(r.URL.Path, bucket)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	if rc != 200 {
		t.Printf("bucket.get failed, head = %v, rc = %d\n",
			head, rc )
		http.Error(w, http.StatusText(rc), rc)   // FIXME, panics
	}
	t.Printf("bucket.get worked, head = %v\n", head)
	for key, value := range head {
		if value != "" {
			w.Header().Set(key, value)
		}
	}
	_, err = w.Write(data)
	if err != nil {
		t.Printf("oopsie! %v\n", err) // FIXME log this
	}

}


