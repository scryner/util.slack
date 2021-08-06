package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/scryner/util.slack/msgfmt"
)

type View struct {
	Type            string            `json:"type"`
	Title           msgfmt.PlainText  `json:"title"`
	Blocks          []msgfmt.Block    `json:"blocks"`
	Close           *msgfmt.PlainText `json:"close,omitempty"`
	Submit          *msgfmt.PlainText `json:"submit,omitempty"`
	PrivateMetadata string            `json:"private_metadata,omitempty"`
	CallbackId      string            `json:"callback_id,omitempty"`
	ClearOnClose    *bool              `json:"clear_on_close,omitempty"`
	NotifyOnClose   *bool              `json:"notify_on_close,omitempty"`
	ExternalId      string            `json:"external_id,omitempty"`
	SubmitDisabled  *bool              `json:"submit_disabled,omitempty"`
}

type genericResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error"`
}

func (api *API) PublishHomeView(email string, blocks []msgfmt.Block) error {
	return api.PublishView(email, &View{
		Type:   "home",
		Title: msgfmt.PlainText{
			Text:  "Hello, world!",
			Emoji: false,
		},
		Blocks: blocks,
	})
}

type publishViewRequest struct {
	UserId string `json:"user_id"`
	View   *View  `json:"view"`
}

func (api *API) PublishView(email string, view *View) error {
	// find user
	user, err := api.SearchUserByEmail(email)
	if err != nil {
		return fmt.Errorf("%w '%s': %v", ErrUserNotFound, email, err)
	}

	req := publishViewRequest{
		UserId: user.ID,
		View:   view,
	}

	// do request
	resp, err := api.doHTTPPostJSON("api/views.publish", nil, req)

	if err != nil {
		return fmt.Errorf("failed to send request to publish view: %v", err)
	}

	defer resp.Body.Close()

	// check result
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to publish view: status = %s", resp.Status)
	}

	// read response body
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to response body: %v", err)
	}

	// unmarshal generic response
	var gresp genericResponse

	err = json.Unmarshal(b, &gresp)
	if err != nil {
		// never reached
		return fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	if !gresp.OK {
		return fmt.Errorf("failed to publish view: %s", gresp.Error)
	}

	return nil
}
