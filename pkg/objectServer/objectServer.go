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
	trace.Trace
	x *log.Logger
}

// New creates a object server
func New(t trace.Trace, x *log.Logger) *Object {
	ceph = cephInterface.New(t, x)
	return &Object{ t, x }
}

// Get an object from a specific bucket. Errors are written to w
func (o Object) Get(w http.ResponseWriter, r *http.Request, bucket string)  { // nolint
	var head map[string]string
	defer o.Begin(r.URL.Path)()

	o.Printf("got a request for %s\n", r.URL.Path)
	data, head, rc, err := ceph.Get(r.URL.Path, bucket)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
	if rc != 200 {
		o.Printf("bucket.get failed, head = %v, rc = %d\n",
			head, rc )
		http.Error(w, http.StatusText(rc), rc)   // FIXME, panics
	}
	o.Printf("bucket.get worked, head = %v\n", head)
	for key, value := range head {
		if value != "" {
			w.Header().Set(key, value)
		}
	}
	_, err = w.Write(data)
	if err != nil {
		o.Printf("oopsie! %v\n", err) // FIXME log this
	}

}


