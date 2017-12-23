package imageServer


// MigrateAndResizeImage gets a file, saves it and calls resize
func (i imager) migrateAndResize(content []byte, key string, width, height, quality uint, grayScale bool, name, imgType string) []byte {
	defer t.Begin(key, width, height, quality, grayScale, name, imgType)()
	return []byte("" )
}
// FIXME wrap this in a check, log on error
