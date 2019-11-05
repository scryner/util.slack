package incominghook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/scryner/util.slack/msgfmt"
)

const (
	DefaultRequestTimeout = 5 * time.Second
)

type Notifier struct {
	webhookURL string
	requestTimeout time.Duration
}

type Option func(*Notifier) error

func RequestTimeout(timeout time.Duration) Option {
	return func(notifier *Notifier) error {
		notifier.requestTimeout = timeout

		return nil
	}
}

func NewNotifier(webhookURL string, opts ...Option) (*Notifier, error) {
	n := &Notifier{
		webhookURL:     webhookURL,
		requestTimeout: DefaultRequestTimeout,
	}

	var err error
	for _, opt := range opts {
		err = opt(n)
		if err != nil {
			return nil, err
		}
	}

	return n, nil
}

func (notifier *Notifier) Notify(message msgfmt.Message) error {
	// marshal message
	b, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal messag: %v", err)
	}

	// make request
	req, err := http.NewRequest(http.MethodPost, notifier.webhookURL, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to make notify request: %v", err)
	}

	ctx,cancel := context.WithTimeout(context.Background(), notifier.requestTimeout)
	defer cancel()

	req = req.WithContext(ctx)

	// do request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to notify request: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failedt to notify request: status = %s", resp.Status)
	}

	return nil
}