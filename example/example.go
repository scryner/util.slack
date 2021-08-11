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
	h := handler{
		slack: slack,
		views: make(map[string]*viewContext),
	}

	s, err := server.New(signingSecret, server.ListenPort(8080),
		server.LogLevel(server.DEBUG),
		server.Handlers(
			server.SlashCommand("/slash", h),
			server.EventSubscriptions("/event", h),
			server.Interactivity("/interactivity", h),
		),
	)

	if err != nil {
		log.Fatalf("failed to make server: %v", err)
	}

	errCh := s.StartServer()

	err = <-errCh
	log.Fatal("failure at server:", err)
}

type viewContext struct {
	viewId  string
	userId  string
	channel string
}

type handler struct {
	slack *api.API
	views map[string]*viewContext
}

func (h handler) HandleCommand(ctx server.Context, req *server.SlashCommandRequest) (msgfmt.Message, error) {
	t := true

	// open modal view
	viewId, err := h.slack.OpenView(req.TriggerId, &api.View{
		Type: "modal",
		Title: msgfmt.PlainText{
			Text:  fmt.Sprintf("Handle '%s' :+1:", req.Text),
			Emoji: true,
		},
		Blocks: []msgfmt.Block{
			msgfmt.Section{
				Text: msgfmt.MarkdownText{
					Text: "Hello modal world!",
				},
			},
			msgfmt.Input{
				Label: msgfmt.PlainText{
					Text: "Title:",
				},
				Element: msgfmt.PlainTextInput{
					ActionId: "input_title",
				},
			},
			msgfmt.Input{
				Label: msgfmt.PlainText{
					Text: "Content:",
				},
				Element: msgfmt.PlainTextInput{
					Multiline: true,
					ActionId:  "input_content",
				},
			},
		},
		Close: &msgfmt.PlainText{
			Text: "Goodbye",
		},
		Submit: &msgfmt.PlainText{
			Text:  "Submit! :heart:",
			Emoji: true,
		},
		NotifyOnClose: &t,
	})

	if err != nil {
		return nil, err
	}

	h.views[viewId] = &viewContext{
		viewId:  viewId,
		userId:  req.UserId,
		channel: req.ChannelId,
	}

	return msgfmt.PlainText{
		Text:  req.Text,
		Emoji: false,
	}, nil
}

func (h handler) HandleEvent(ctx server.Context, cb *server.EventCallback) error {
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

func (h handler) HandleBlockActions(ctx server.Context, blockActions *server.BlockActions) error {
	panic("implement me")
}

func (h handler) HandleMessageActions(ctx server.Context, messageActions *server.MessageActions) error {
	panic("implement me")
}

func (h handler) HandleViewClosed(ctx server.Context, viewClosed *server.ViewClosed) error {
	b, _ := json.MarshalIndent(viewClosed, "", "  ")
	fmt.Println(string(b))

	return nil
}

func (h handler) HandleViewSubmission(ctx server.Context, viewSubmission *server.ViewSubmission) error {
	viewId, _ := viewSubmission.View["id"].(string)
	vctx, ok := h.views[viewId]
	if !ok {
		return fmt.Errorf("failed to get view '%s'", viewId)
	}

	title, _ := viewSubmission.State.GetValue("input_title")
	content, _ := viewSubmission.State.GetValue("input_content")

	msg := fmt.Sprintf("%s: %v", title, content)

	if _, err := h.slack.PostMessage(vctx.channel, &api.ChatMessage{
		Text: msg,
		Blocks: []msgfmt.Block{
			msgfmt.Section{
				Text: msgfmt.MarkdownText{
					Text: msg,
				},
			},
		},
	}); err != nil {
		return err
	}

	return nil
}
