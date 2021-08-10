package main

import (
	"encoding/json"
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

	// make api handle
	slack, err := api.New(accessToken)
	if err != nil {
		log.Fatal("failed to make api:", err)
	}

	// search user
	user, err := slack.SearchUserByEmail("scryner@42dot.ai")
	if err != nil {
		log.Fatal("failed to find user:", err)
	}

	// publish home view
	err = slack.PublishHomeView(user, []msgfmt.Block{
		msgfmt.Section{
			Text: msgfmt.PlainText{
				Text: "My sweet home",
			},
		},
	})

	if err != nil {
		log.Fatal("failed to update home:", err)
	}

	// create server
	s, err := server.New(signingSecret, server.ListenPort(8080),
		server.LogLevel(server.DEBUG),
		server.Handlers(
			server.SlashCommand("/slash", cmdHandler{}),
			server.EventSubscriptions("/event", eventHandler{slack: slack}),
			server.Interactivity("/interactivity", interactivityHandler{slack: slack}),
		),
	)

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

func (h eventHandler) HandleEvent(ctx server.Context, cb *server.EventCallback) error {
	b, _ := json.MarshalIndent(cb, "", "  ")
	fmt.Println(string(b))

	ev := cb.Event

	// get type
	typ, subType, err := ev.Type()
	if err != nil {
		return err
	}

	switch typ {
	case "app_home_opened":
		return nil
	case "message":
		// check message where from; messages from bot are to be ignored
		if _, ok := ev["bot_id"]; ok {
			// pass
			return nil
		}

		switch subType {
		// ignore it
		case "message_deleted":
			fallthrough
		case "bot_message":
			return nil
		}

		userId, _ := ev["user"].(string)
		text, _ := ev["text"].(string)
		//timestamp, _ := ev["ts"].(string)

		// get user info
		user, err := h.slack.GetUserInfo(userId)
		if err != nil {
			return err
		}

		// post echo message
		toBeDelChannel, toBeDelTs, err := h.slack.PostBotDirectMessage(user, &api.ChatMessage{
			Text: text,
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
		if err = h.slack.DeleteMessage(toBeDelChannel, toBeDelTs); err != nil {
			return err
		}

		return nil

	default:
		return fmt.Errorf("unknown typ '%s'", typ)
	}
}

type interactivityHandler struct {
	slack *api.API
}

func (h interactivityHandler) HandleBlockActions(ctx server.Context, blockActions *server.BlockActions) error {
	panic("implement me")
}

func (h interactivityHandler) HandleMessageActions(ctx server.Context, messageActions *server.MessageActions) error {
	panic("implement me")
}

func (h interactivityHandler) HandleViewClosed(ctx server.Context, viewClosed *server.ViewClosed) error {
	panic("implement me")
}

func (h interactivityHandler) HandleViewSubmission(ctx server.Context, viewSubmission *server.ViewSubmission) error {
	panic("implement me")
}
