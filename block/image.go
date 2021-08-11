package block

import (
	"encoding/json"
)

type ImageWithTitle struct {
	Title    PlainText
	ImageUrl string
	AltText  string
}

func (img ImageWithTitle) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":      "image",
		"title":     img.Title,
		"image_url": img.ImageUrl,
		"alt_text":  img.AltText,
	})
}

func (ImageWithTitle) blockAble() {}

type Image struct {
	ImageUrl string
	AltText  string
}

func (image Image) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":      "image",
		"image_url": image.ImageUrl,
		"alt_text":  image.AltText,
	})
}

func (Image) blockAble()            {}
func (Image) contextElementAble()   {}
func (Image) sectionAccessoryAble() {}
