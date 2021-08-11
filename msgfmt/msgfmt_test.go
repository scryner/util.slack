package msgfmt

import (
	"encoding/json"
	"testing"
)

func TestText(t *testing.T) {
	plain := PlainText{
		Text:  "hello, world",
		Emoji: true,
	}

	b, err := json.MarshalIndent(plain, "", "  ")
	if err != nil {
		t.Error("failed to marshal plain text:", err)
		t.FailNow()
	}

	t.Log("plain =>", string(b))

	markdown := MarkdownText{
		Text: "This is markdown text, <http://google.com|this is a link>",
	}

	b, err = json.MarshalIndent(markdown, "", "  ")
	if err != nil {
		t.Error("failed to marshal markdown text:", err)
		t.FailNow()
	}

	t.Log("markdown =>", string(b))
}

func TestBlock(t *testing.T) {
	var blocks Blocks

	// section
	blocks = append(blocks, Section{
		Text: MarkdownText{
			Text: "This is a section",
		},
	})

	// section with image
	blocks = append(blocks, Section{
		Text: MarkdownText{
			Text: "this is a section with image",
		},
		Accessory: ImageElement{
			ImageURL: "https://api.slack.com/img/blocks/bkb_template_images/palmtree.png",
			AltText:  "palm tree",
		},
	})

	// image
	blocks = append(blocks, Image{
		Title: PlainText{
			Text:  "Example image",
			Emoji: true,
		},
		ImageURL: "https://api.slack.com/img/blocks/bkb_template_images/goldengate.png",
		AltText:  "Example Image",
	})

	// context
	blocks = append(blocks, Context{
		Elements: []SectionAccessory{
			MarkdownText{
				Text: "this is a markdown element in context",
			},
			ImageElement{
				ImageURL: "https://api.slack.com/img/blocks/bkb_template_images/palmtree.png",
				AltText:  "palm tree",
			},
		},
	})

	// divider
	blocks = append(blocks, Divider())

	// section with button
	blocks = append(blocks, Section{
		Text: MarkdownText{
			Text: "this is a section with button",
		},
		Accessory: Button{
			Text: PlainText{
				Text: "Press Me!",
			},
			URL: "press123",
		},
	})

	// marshal
	b, err := json.MarshalIndent(blocks, "", "  ")
	if err != nil {
		t.Error("failed to marshal blocks:", err)
		t.FailNow()
	}

	t.Log("blocks =>", string(b))
}
