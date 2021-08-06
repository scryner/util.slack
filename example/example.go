package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/scryner/util.slack/api"
	"github.com/scryner/util.slack/msgfmt"
	"github.com/scryner/util.slack/server"
)

func main() {
	accessToken := os.Getenv("ACCESS_TOKEN")     // bot token
	signingSecret := os.Getenv("SIGNING_SECRET") // signing token

	slack, err := api.New(accessToken)
	if err != nil {
		log.Fatal("failed to make api:", err)
	}

	s, err := server.New(signingSecret, server.ListenPort(8080),
		server.LogLevel(server.DEBUG),
		server.Handlers(
			server.SlashCommand("/slash", cmdHandler{}),
			server.EventSubscriptions("/event",
				server.EventHandlerDef("event_callback", eventHandler{
					slack: slack,
				}))))

	if err != nil {
		log.Fatalf("failed to make server: %v", err)
	}

	errCh := s.StartServer()

	err = <-errCh
	log.Fatal("failure at server:", err)
}

type cmdHandler struct{}

func (cmdHandler) HandleCommand(ctx server.Context, req *server.SlashCommandRequest) (msgfmt.Message, error) {
	return msgfmt.PlainText{
		Text:  req.Text,
		Emoji: false,
	}, nil
}

type eventHandler struct {
	slack *api.API
}

func (h eventHandler) HandleEvent(ctx server.Context, prop map[string]interface{}) error {
	b, _ := json.MarshalIndent(prop, "", "  ")
	fmt.Println(string(b))

	ev, ok := prop["event"].(map[string]interface{})
	if !ok {
		return errors.New("event property not found")
	}

	// check message where from; messages from bot are to be ignored
	subType, _ := ev["subtype"].(string)
	switch subType {
	// ignore it
	case "message_deleted":
		fallthrough
	case "bot_message":
		return nil
	}

	channel, _ := ev["channel"].(string)
	userId, _ := ev["user"].(string)
	text, _ := ev["text"].(string)
	//timestamp, _ := ev["ts"].(string)

	// get user info
	user, err := h.slack.GetUserInfo(userId)
	if err != nil {
		return err
	}

	// post echo message
	toBeDelChannel, toBeDelTs, err := h.slack.PostMessage(user.Profile.Email, &api.ChatMessage{
		ChannelID:        channel,
		NotificationText: "echo",
		Blocks: []msgfmt.Block{msgfmt.Section{
			Text: msgfmt.PlainText{
				Text: text,
			},
		}},
	})

	if err != nil {
		return err
	}

	// delete message after 3 seconds
	time.Sleep(time.Second * 3)
	if err = h.slack.DeleteMessage(toBeDelChannel, toBeDelTs, false); err != nil {
		return err
	}

	return nil
}
