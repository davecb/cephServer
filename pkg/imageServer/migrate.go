package imageServer


// MigrateAndResizeImage gets a file, savwes it and calls resize
func MigrateAndResizeImage(content, key string, width, height, quality uint, grayScale bool, name, imgType string) string {
	defer T.Begin(key, width, height, quality, grayScale, name, imgType)()
	return ""
}
