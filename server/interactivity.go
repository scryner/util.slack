package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
)

type User struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	TeamId   string `json:"team_id"`
}

type Team struct {
	Id     string `json:"id"`
	Domain string `json:"domain"`
}

type Channel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Action struct {
	Type     string `json:"type"`
	BlockId  string `json:"block_id"`
	ActionId string `json:"action_id"`
	Value    string `json:"value"`
	Content  map[string]interface{}
}

func (a *Action) UnmarshalJSON(b []byte) error {
	action := Action{
		Content: make(map[string]interface{}),
	}

	var m map[string]interface{}

	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	for k, v := range m {
		switch k {
		case "type":
			action.Type = safeToString(v)
		case "block_id":
			action.BlockId = safeToString(v)
		case "action_id":
			action.ActionId = safeToString(v)
		case "value":
			action.Value = safeToString(v)
		default:
			action.Content[k] = v
		}
	}

	*a = action
	return nil
}

func safeToString(v interface{}) string {
	s, _ := v.(string)
	return s
}

type BlockActions struct {
	TriggerId   string                 `json:"trigger_id"`
	ResponseUrl string                 `json:"response_url"`
	User        User                   `json:"user"`
	Team        Team                   `json:"team"`
	Message     map[string]interface{} `json:"message"`
	View        map[string]interface{} `json:"view"`
	Actions     []Action               `json:"actions"`
	Hash        string                 `json:"hash"`
}

type Message struct {
	Type string `json:"type"`
	User string `json:"user"`
	Ts   string `json:"ts"`
	Text string `json:"text"`
}

type MessageActions struct {
	CallbackId  string  `json:"callback_id"`
	TriggerId   string  `json:"trigger_id"`
	ResponseUrl string  `json:"response_url"`
	User        User    `json:"user"`
	Message     Message `json:"message"`
	Channel     Channel `json:"channel"`
	Team        Team    `json:"team"`
}

type ViewClosed struct {
	Team      Team                   `json:"team"`
	User      User                   `json:"user"`
	View      map[string]interface{} `json:"view"`
	IsCleared bool                   `json:"is_cleared"`
}

type ResponseUrl struct {
	BlockId     string `json:"block_id"`
	ActionId    string `json:"action_id"`
	ChannelId   string `json:"channel_id"`
	ResponseUrl string `json:"response_url"`
}

type ViewSubmission struct {
	Team         Team                   `json:"team"`
	User         User                   `json:"user"`
	View         map[string]interface{} `json:"view"`
	Hash         string                 `json:"hash"`
	ResponseUrls []ResponseUrl          `json:"response_urls"`
}

type InteractivityHandler interface {
	HandleBlockActions(ctx Context, blockActions *BlockActions) error
	HandleMessageActions(ctx Context, messageActions *MessageActions) error
	HandleViewClosed(ctx Context, viewClosed *ViewClosed) error
	HandleViewSubmission(ctx Context, viewSubmission *ViewSubmission) error
}

func Interactivity(endpoint string, handler InteractivityHandler) handler {
	return func() (string, echo.HandlerFunc) {
		return endpoint, func(ctx echo.Context) error {
			// get request body
			reqBody, ok := ctx.Get("reqBody").([]byte)
			if !ok {
				// actually, never happened
				ctx.Logger().Errorf("empty request body")
				return ctx.JSON(http.StatusBadRequest, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't read your request body",
				})
			}

			// get payload
			formVals, err := url.ParseQuery(string(reqBody))
			if err != nil {
				ctx.Logger().Errorf("failed to read form param: %v", err)
				return ctx.JSON(http.StatusBadRequest, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't read your form params",
				})
			}

			payload := []byte(formVals.Get("payload"))
			if payload == nil {
				ctx.Logger().Errorf("empty payload param")
				return ctx.JSON(http.StatusBadRequest, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't read your form params: empty",
				})
			}

			// unmarshal to map
			var props map[string]interface{}

			err = json.Unmarshal(payload, &props)
			if err != nil {
				ctx.Logger().Errorf("failed to unmarshal request to json: %v", err)
				return ctx.JSON(http.StatusBadRequest, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't understand your request body",
				})
			}

			b, _ := json.MarshalIndent(props, "", "  ")
			fmt.Println(string(b))

			// get event type
			typ, ok := props["type"].(string)
			if !ok || typ == "" {
				ctx.Logger().Errorf("event type was nil")
				return ctx.JSON(http.StatusBadRequest, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't determine event type",
				})
			}

			// dispatch payload
			switch typ {
			case "url_verification":
				challenge, _ := props["challenge"].(string)
				return ctx.String(http.StatusOK, challenge)

			case "block_actions":
				// unmarshal payload
				var blockActions BlockActions
				if err := json.Unmarshal(payload, &blockActions); err != nil {
					ctx.Logger().Errorf("failed to unmarshal block actions to json: %v", err)
					return ctx.JSON(http.StatusBadRequest, slackError{
						ResponseType: "ephemeral",
						Text:         "I can't understand your block actions payload",
					})
				}

				// dispatch it
				if handler.HandleBlockActions != nil {
					if err := handler.HandleBlockActions(ctx, &blockActions); err != nil {
						ctx.Logger().Errorf("failed to handle block actions: %v", err)
						return ctx.JSON(http.StatusInternalServerError, slackError{
							ResponseType: "ephemeral",
							Text:         "I can't handle your block actions",
						})
					}
				}

			case "message_actions":
				var messageActions MessageActions
				if err := json.Unmarshal(payload, &messageActions); err != nil {
					ctx.Logger().Errorf("failed to unmarshal message actions to json: %v", err)
					return ctx.JSON(http.StatusBadRequest, slackError{
						ResponseType: "ephemeral",
						Text:         "I can't understand your message actions payload",
					})
				}

				// dispatch it
				if handler.HandleMessageActions != nil {
					if err := handler.HandleMessageActions(ctx, &messageActions); err != nil {
						ctx.Logger().Errorf("failed to handle message actions: %v", err)
						return ctx.JSON(http.StatusInternalServerError, slackError{
							ResponseType: "ephemeral",
							Text:         "I can't handle your message actions",
						})
					}
				}

			case "view_closed":
				var viewClosed ViewClosed
				if err := json.Unmarshal(payload, &viewClosed); err != nil {
					ctx.Logger().Errorf("failed to unmarshal view closed to json: %v", err)
					return ctx.JSON(http.StatusBadRequest, slackError{
						ResponseType: "ephemeral",
						Text:         "I can't understand your view closed payload",
					})
				}

				// dispatch it
				if handler.HandleViewClosed != nil {
					if err := handler.HandleViewClosed(ctx, &viewClosed); err != nil {
						ctx.Logger().Errorf("failed to handle view closed: %v", err)
						return ctx.JSON(http.StatusInternalServerError, slackError{
							ResponseType: "ephemeral",
							Text:         "I can't handle your view closed",
						})
					}
				}

			case "view_submission":
				var viewSubmission ViewSubmission
				if err := json.Unmarshal(payload, &viewSubmission); err != nil {
					ctx.Logger().Errorf("failed to unmarshal view submission to json: %v", err)
					return ctx.JSON(http.StatusBadRequest, slackError{
						ResponseType: "ephemeral",
						Text:         "I can't understand your view submission payload",
					})
				}

				// dispatch it
				if handler.HandleViewSubmission != nil {
					if err := handler.HandleViewSubmission(ctx, &viewSubmission); err != nil {
						ctx.Logger().Errorf("failed to handle view submission: %v", err)
						return ctx.JSON(http.StatusInternalServerError, slackError{
							ResponseType: "ephemeral",
							Text:         "I can't handle your view submission",
						})
					}
				}

			default:
				ctx.Logger().Errorf("unknown type '%s'", typ)
				return ctx.JSON(http.StatusBadRequest, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't know interactivity payload type",
				})
			}

			return nil
		}
	}
}
