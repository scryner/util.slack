package msgfmt

import (
	"encoding/json"
)

type Image struct {
	Title    Text
	ImageURL string
	AltText  string
}

func (img Image) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type":      "image",
		"title":     img.Title,
		"image_url": img.ImageURL,
		"alt_text":  img.AltText,
	}

	return json.Marshal(m)
}

func (Image) blockAble() {}

type ImageElement struct {
	ImageURL string
	AltText  string
}

func (img ImageElement) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{
		"type":      "image",
		"image_url": img.ImageURL,
		"alt_text":  img.AltText,
	}

	return json.Marshal(m)
}

func (ImageElement) elementAble() {}
