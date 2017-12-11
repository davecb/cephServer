package main

import (
	"fmt"
	"strconv"
	"strings"
)

// handle the imager-specific parsing problem
func parseImageURL(s string) (key string, width, height, quality uint,
	grayScale bool, name, imgType string, err error) {
	const defaultQuality = 85

	defer T.Begin()()
	tokens := strings.Split(s, "/")
	at := len(tokens) - 1
	T.Printf("tokens = %v\n", tokens)
	if at <= 0 {
		// FIXME this may be acceptable at a later time
		return "", 0, 0, 0, false, "", "",
			fmt.Errorf("could not find any / characters in %q, rejected", s)
	}

	// Proceed from right to left, although this is LL(1)
	T.Printf("name.type token[%d] = %q\n", at, tokens[at])
	at, name, imgType = parseNameComponent(tokens, at)

	// We are now before the name, expecting a boolean, a number or a text key
	T.Printf("quality token[%d] = %q\n", at, tokens[at])
	at, grayScale = parseGrayscale(tokens, at)

	// we are now past (sorta) grayScale, expecting a quality,
	// a height, a width or a text key , in that order
	T.Printf("quality token[%d] = %q\n", at, tokens[at])
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
	T.Printf("height token[%d] = %q\n", at, tokens[at])
	u, err = strconv.ParseUint(tokens[at], 10, 64)
	if err != nil {
		// height is dissed by the imager, so 0 is ok
		height = -0
	} else {
		height = uint(u)
		at = decrement(at)
	}

	// Headed toward just width and key
	T.Printf("width token[%d] = %q\n", at, tokens[at])
	u, err = strconv.ParseUint(tokens[at], 10, 64)
	if err != nil {
		width = 0
	} else {
		width = uint(u)
		at = decrement(at)
	}

	// OK, anything else is key, even if it has slashes in it
	T.Printf("key token[%d] = %q\n", at, tokens[at])
	for i := 0; i <= at; i++ {
		if key == "" {
			key = tokens[i]
		} else {
			key = key + "/" + tokens[i]
		}
	}
	T.Printf("key = %q\n", key)
	return key, width, height, quality, grayScale, name, imgType, nil
}


// parseNameComponent parses strings like name.gif
func parseNameComponent(tokens []string, at int) (int, string, string) {
	var imgType string
	
	nameAndType := strings.Split(tokens[at], ".")
	switch nameAndType[1] {
	case "jpg", "jpeg", "JPG", "JPEG", "png", "PNG":  // Webp? likely
		imgType = nameAndType[1]
	default:
		imgType = "" // no type
	}
	return decrement(at), nameAndType[0], imgType
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

