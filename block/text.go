package block

import (
	"encoding/json"
)

type Text interface {
	text() string
	Message
}

type PlainText struct {
	Text  string
	Emoji bool
}

func (t PlainText) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type":  "plain_text",
		"text":  t.Text,
		"emoji": t.Emoji,
	})
}

func (t PlainText) text() string {
	return t.Text
}

func (PlainText) sendAble()             {}
func (PlainText) sectionAccessoryAble() {}
func (PlainText) contextElementAble()   {}

type MarkdownText struct {
	Text string
}

func (t MarkdownText) text() string {
	return t.Text
}

func (t MarkdownText) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type": "mrkdwn",
		"text": t.Text,
	})
}

func (MarkdownText) sendAble()             {}
func (MarkdownText) sectionAccessoryAble() {}
func (MarkdownText) contextElementAble()   {}
