package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/scryner/util.slack/api"
	"github.com/scryner/util.slack/block"
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
	err = slack.PublishHomeView(user, []block.Block{
		block.Section{
			Text: block.PlainText{
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
	}

	s, err := server.New(signingSecret, server.ListenPort(8080),
		server.LogLevel(server.DEBUG),
		server.Handlers(
			server.SlashCommand("/slash", h),
			server.EventSubscriptions("/event", h),
			server.Interactivity("/interactivity", h),
			server.Http(http.MethodPost, "/echo", func(ctx echo.Context) error {
				// read body
				b, err := ioutil.ReadAll(ctx.Request().Body)
				if err != nil {
					return ctx.String(http.StatusBadRequest, "can't read request body")
				}

				if len(b) < 1 {
					return ctx.String(http.StatusBadRequest, "empty request body")
				}

				return ctx.String(http.StatusOK, string(b))
			}),
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
}

func (h handler) HandleCommand(ctx server.Context, req *server.SlashCommandRequest) (block.Message, error) {
	t := true

	// open modal view
	_, err := h.slack.OpenView(req.TriggerId, &api.View{
		Type: "modal",
		Title: block.PlainText{
			Text:  fmt.Sprintf("Handle '%s' :+1:", req.Text),
			Emoji: true,
		},
		Blocks: []block.Block{
			block.Section{
				Text: block.MarkdownText{
					Text: "Hello modal world!",
				},
			},
			block.Input{
				Label: block.PlainText{
					Text: "Title:",
				},
				Element: block.PlainTextInput{
					ActionId: "input_title",
					PlaceHolder: block.PlainText{
						Text:  "Hello!",
					},
				},
			},
			block.Input{
				Label: block.PlainText{
					Text: "Content:",
				},
				Element: block.PlainTextInput{
					Multiline: true,
					ActionId:  "input_content",
				},
			},
		},
		Close: &block.PlainText{
			Text: "Goodbye",
		},
		Submit: &block.PlainText{
			Text:  "Submit! :heart:",
			Emoji: true,
		},
		NotifyOnClose:   &t,
		PrivateMetadata: []byte(req.ChannelId),
	})

	if err != nil {
		return nil, err
	}

	return block.PlainText{
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
			Blocks: []block.Block{block.Section{
				Text: block.PlainText{
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
	channel := string(viewSubmission.View.GetPrivateMetadata())
	title := viewSubmission.State.GetValue("input_title")
	content := viewSubmission.State.GetValue("input_content")

	msg := fmt.Sprintf("(%s) %s: %v", channel, title, content)
	fmt.Println(msg)

	return h.slack.PostEphemeralMessage(channel, viewSubmission.User.Id, &api.ChatMessage{
		Text: msg,
		Blocks: block.Blocks{
			block.Section{
				Text: block.PlainText{
					Text: msg,
				},
			},
		},
	})
}
