package imageServer

import (
	"github.com/davecb/cephServer/pkg/cephInterface"
	"github.com/davecb/cephServer/pkg/trace"

	"fmt"
	"strconv"
	"net/http"
	"strings"
	"log"
)

const largestWidth = 100
var ceph *cephInterface.S3Proto
var t trace.Trace

// imager is a resizing mechanism
type imager struct {
	trace.Trace
	logger *log.Logger
}


// New creates an image-resizer
func New(x trace.Trace, y *log.Logger) *imager {
	ceph = cephInterface.New(x, y)
	t = x // this is subtly odd.
	return &imager{ x,y }
}


// GetSized gets an image in a specific resize, using specific buckets
func (i imager) GetSized(w http.ResponseWriter, r *http.Request) {
	defer t.Begin(r.URL.Path)()
	downloadBucket := "download.s3.kobo.com"
	imageBucket := "images.s3.kobo.com"

	fullPath := r.URL.Path
	// first, do a prerequisite
	key, width, height, quality, grey, name, imgType, err := parseImageURL(fullPath)
	if err != nil {
		http.Error(w, "Cannot interpret url " + fullPath , 400)
		return
	}

	// next, see if we have exactly the image asked for
	bytes, head, rc, err := ceph.Get(fullPath, imageBucket)
	if err == nil && rc == 200 {
		// return the file in the size requested
		t.Printf("found the full path in imager, head = %v\n", head)
		w.Write(bytes) // nolint ignore error???  FIXME
		return
	}
	t.Printf("didn't find the full path %s, trying for base\n", fullPath)
	// postcondition: we didn't find the image, so ...


	// go looking for the base (big) version to resize. Note the images
	// have a prefix and the downloads don't. This is different from
	// mogilefs, which saved both with an images/prefix.
	bytes, head, rc, err = ceph.Get("images/" + key, downloadBucket)
	if err == nil && rc == 200 {
		t.Printf("found the base in imager, head = %v\n", head)
		// we have a base file which we can resize
		if width < largestWidth {
			// we can afford to do it in-line
			t.Printf("going to resize in-line\n")
			b := i.resize(bytes, key, width, height,
				quality, grey, name, imgType)
			// return it, and save in the background
			write(w, b)
			go ceph.Put(b, fullPath, imageBucket) // nolint
		} else {
			// we background it and return either a dummy FIXME or the original
			t.Printf("going to resize in background\n")
			write(w, bytes)
			go func() {
				ceph.Put(i.resize(bytes, key, // nolint
					width, height, quality, grey, name, imgType), fullPath, imageBucket)
			}()
		}
		return
	}
	t.Printf("didn't find the base, don't try for a migration, just fail\n")
	// FIXME log this
	http.Error(w, "file not found, no master to create from", 404)

	// Alternate teatment: we lack a base, so background a
	// migrate-then-resize operation and return a dummy
	//write(w, i.getDummy(imgType)) // nolint
	//go func() {
	//	ceph.Put(i.migrateAndResize(bytes, // nolint
	//		key, width, height, quality, grey, name, imgType), fullPath, imageBucket)
	//}()
}

// write logs write errors
func write(w http.ResponseWriter, b []byte) {
	_, err := w.Write(b)
	if err != nil {
		// log it
	}
}


// handle the imager-specific parsing problem
func parseImageURL(s string) (key string, width, height, quality uint,
	grayScale bool, name, imgType string, err error) {
	const defaultQuality = 85

	defer t.Begin()()
	tokens := strings.Split(s, "/")
	at := len(tokens) - 1
	t.Printf("tokens = %v\n", tokens)
	if at <= 0 {
		// FIXME this may be acceptable at a later time
		// FIXME explain the error in the log
		return "", 0, 0, 0, false, "", "",
			fmt.Errorf("could not find any / characters in %q, rejected", s)
	}
	// FIXME accept "images/<key>", too

	// Proceed from right to left, although this is LL(1)
	t.Printf("name.type token[%d] = %q\n", at, tokens[at])
	at, name, imgType = parseNameComponent(tokens, at)

	// We are now before the name, expecting a boolean, a number or a text key
	t.Printf("quality token[%d] = %q\n", at, tokens[at])
	at, grayScale = parseGrayscale(tokens, at)

	// we are now past (sorta) grayScale, expecting a quality,
	// a height, a width or a text key , in that order
	t.Printf("quality token[%d] = %q\n", at, tokens[at])
	u, err := strconv.ParseUint(tokens[at], 10, 64)
	if err != nil {
		// not a number, and the returned value is 0,
		// or to big a value, and the number is trash
		quality = defaultQuality
	} else if u > 100 {
		// its a number, but too big to be a quality
		quality = defaultQuality
	} else {
		quality = uint(u)
		at = decrement(at)
	}

	// We are sorta past quality, looking for one of
	// a height, a width or a text key
	t.Printf("height token[%d] = %q\n", at, tokens[at])
	u, err = strconv.ParseUint(tokens[at], 10, 64)
	if err != nil {
		// height is dissed by the imager, so 0 is ok
		height = -0
	} else {
		height = uint(u)
		at = decrement(at)
	}

	// Headed toward just width and key
	t.Printf("width token[%d] = %q\n", at, tokens[at])
	u, err = strconv.ParseUint(tokens[at], 10, 64)
	if err != nil {
		width = 0
	} else {
		width = uint(u)
		at = decrement(at)
	}

	// OK, anything else is key, even if it has slashes in it
	t.Printf("key token[%d] = %q\n", at, tokens[at])
	for j := 0; j <= at; j++ {
		if key == "" {
			key = tokens[j]
		} else {
			key = key + "/" + tokens[j]
		}
	}
	t.Printf("key = %q\n", key)
	return key, width, height, quality, grayScale, name, imgType, nil
}


// parseNameComponent parses strings like name.gif
func parseNameComponent(tokens []string, at int) (int, string, string) {
	var imgType string


	nameAndType := strings.Split(tokens[at], ".")
	if len(nameAndType) == 1 {
		var name, imgType string
		s :=  tokens[at]
		switch {
		case strings.HasSuffix(s, "."):
			imgType = ""
			name = strings.TrimSuffix(s, ".")
		case strings.HasPrefix(s, "."):
			imgType = strings.TrimPrefix(s, ".")
			name = ""
		case imageType(s) != "":
			imgType = s
			name = ""
		default:
			imgType = ""
			name = s
		}
		return decrement(at), name, imgType
	}
	imgType = imageType(nameAndType[1])
	return decrement(at), nameAndType[0], imgType
}

// imaheType accept the few types we allow
func imageType(s string) string {
	switch s {
	case "jpg", "jpeg", "JPG", "JPEG", "png", "PNG": // Webp? likely
		return s
	default:
		return "" // no type
	}
}

// parseGrayscale sees if this token is a grayscale true or false
func parseGrayscale(tokens []string, at int) (int, bool) {
	var  grayScale bool

	switch tokens[at] {
	case "true", "True":
		grayScale = true
		at = decrement(at)
	case "false", "False", "":
		grayScale = false
		at = decrement(at)
	default:
		// if we get here, we lack a grayScale value
		grayScale = false
	}
	return at, grayScale
}


// decrement a counter toward zero, but no lower
func decrement(i int) int {
	if i == 0 {
		return 0
	}
	return i - 1
}


// return a dummy image in the appropriate type and a selected resize
func (i imager) getDummy(imageType string) []byte {
	defer t.Begin()()
	return []byte("dummy image")
}

