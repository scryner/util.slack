package incominghook

import (
	"os"
	"testing"

	"github.com/scryner/util.slack/block"
)

func TestNotifier(t *testing.T) {
	// try to get url from os env
	webhookURL := os.Getenv("WEBHOOK_URL")

	notifier, err := NewNotifier(webhookURL)
	if err != nil {
		t.Error("failed to make notifier:", err)
		t.FailNow()
	}

	// notify a message
	err = notifier.Notify(block.PlainText{
		Text:  "Wake up! You're only hope.",
	})

	if err != nil {
		t.Error("failed to notify:", err)
		t.FailNow()
	}
}