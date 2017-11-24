package cephInterface

import (
	"imageServer/pkg/trace"
	"os"
	"log"
	"strings"
)

// T is a debugging tool shared by the server components
var T trace.Trace

// Get obtains a file from ceph via the s3-compatible interface
func Get(key string) string {
	defer T.Begin(key)()
	return ""
}

// WeHave sees if we have a file by its key
// it may be follwed by a Get if it returns true
func WeHave(key string) bool {
	defer T.Begin(key)()

	switch {
	case key == "image/albert/100/200/85/False/albert.jpg":
		return false
	case key == "image/albert":
		return true
	default:
		return false
	}
}

// Save stores a file in ceph via the s3-compatible interface
func Save(key, contents string) {
	defer T.Begin(key)()

	strings := strings.Split(key, "/")
	out, err := os.Create(strings[len(strings)-1])
	if err != nil {
		log.Fatal(err)
	}
	_, err = out.WriteString(contents)
	if err != nil {
		log.Fatalf("write failure %v\n", err)
	}
	err = out.Close()
	if err != nil {
		log.Fatalf("write failure %v\n", err)
	}
	
}
