package imageServer


// MigrateAndResizeImage gets a file, savwes it and calls resize
func (image Imager) migrateAndResize(content []byte, key string, width, height, quality uint, grayScale bool, name, imgType string) string {
	defer image.Begin(key, width, height, quality, grayScale, name, imgType)()
	return ""
}
// FIXME wrap this in a check, log on error
