package imageMigrator

import "imageServer/pkg/trace"

// T is a debugging tool shared by the server components
var T trace.Trace  // a debugging tool

// MigrateAndResizeImage gets a file, savwes it and calls resize
func MigrateAndResizeImage(key string, width, height, quality uint, grayScale bool, name, imgType string) string {
	defer T.Begin(key, width, height, quality, grayScale, name, imgType)()
	return ""
}
