package slashcmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"

	"github.com/scryner/util.slack/msgfmt"
)

type LogLvl uint8

const (
	DEBUG LogLvl = iota + 1
	INFO
	WARN
	ERROR
)

const (
	DefaultListenPort = 8080
)

type slackError struct {
	ResponseType string `json:"response_type"`
	Text         string `json:"text"`
}

type Server struct {
	listenPort    int
	signingSecret string
	logLevel      log.Lvl
	handler       CommandHandler
}

type Request struct {
	ChannelID   string
	ChannelName string
	Command     string
	ResponseURL string
	TeamDomain  string
	TeamID      string
	Text        string
	Token       string
	TriggerID   string
	UserID      string
	UserName    string
}

type CommandHandler interface{
	Handle(*Request) (msgfmt.Message, error)
}

type Option func(*Server) error

func NewServer(signingSecret string, opts ...Option) (*Server, error) {
	srv := &Server{
		listenPort:    DefaultListenPort,
		signingSecret: signingSecret,
		logLevel:      log.INFO,
	}

	// apply options
	var err error
	for _, opt := range opts {
		err = opt(srv)
		if err != nil {
			return nil, err
		}
	}

	return srv, nil
}

func ListenPort(port int) Option {
	return func(server *Server) error {
		server.listenPort = port

		return nil
	}
}

func LogLevel(level LogLvl) Option {
	return func(server *Server) error {
		server.logLevel = log.Lvl(level)

		return nil
	}
}

func Handler(handler CommandHandler) Option {
	return func(server *Server) error {
		server.handler = handler

		return nil
	}
}

func (server *Server) StartServer() <-chan error {
	// make error chan
	errCh := make(chan error, 1)

	go func() {
		// make echo
		e := echo.New()
		e.Use(middleware.Logger())
		e.Logger.SetLevel(server.logLevel)

		if server.signingSecret == "" {
			errCh <- errors.New("empty signing secret")
			return
		}

		// make verifier
		verifier := NewVerifier(server.signingSecret)

		// register handler
		e.POST("/*", func(ctx echo.Context) error {
			// verify token
			reqTimestamp := fromHeaderAsInt64(ctx.Request().Header, "X-Slack-Request-Timestamp")
			reqSignature := ctx.Request().Header.Get("X-Slack-Signature")

			reqBody, err := ioutil.ReadAll(ctx.Request().Body)
			if err != nil {
				ctx.Logger().Errorf("failed to read request body: %v", err)
				return ctx.JSON(http.StatusBadRequest, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't read your request body",
				})
			}

			err = verifier.Verify(reqTimestamp, reqSignature, string(reqBody))
			if err != nil {
				ctx.Logger().Errorf("failed to verify request: %v", err)
				return ctx.JSON(http.StatusForbidden, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't verify your request body",
				})
			}

			/*
				# form raw body
				token=zFqDfA3ODOYHlGXSE1jn3DYT&team_id=THABEEB7X&team_domain=code42ai&channel_id=CHH2RECA1&channel_name=engineering&user_id=UHH47S7HB&user_name=scryner&command=%2Flight&text=503+off&response_url=https%3A%2F%2Fhooks.slack.com%2Fcommands%2FTHABEEB7X%2F589257081155%2FqolNuJgYhx4jzj2kRMf42k2c&trigger_id=594457234865.588388487269.9d1d657fe0def716ea7497ee36769159"

				# form example
				channel_id:[CH19XRPGR]
				channel_name:[random]
				command:[/light]
				response_url:[https://hooks.slack.com/commands/THABEEB7X/603598559413/GJt9wF1Wp33mStAXSfvMFAAS]
				team_domain:[code42ai]
				team_id:[THABEEB7X]
				text:[503 100]
				token:[zFqDfA3ODOYHlGXSE1jn3DYT]
				trigger_id:[594384164881.588388487269.2f493ca6841a0033fc2ec351c258081a]
				user_id:[UHH47S7HB]
				user_name:[scryner]
			*/

			// get command value
			formVals, err := url.ParseQuery(string(reqBody))
			if err != nil {
				ctx.Logger().Errorf("failed to read form param: %v", err)
				return ctx.JSON(http.StatusBadRequest, slackError{
					ResponseType: "ephemeral",
					Text:         "I can't read your form params",
				})
			}

			request := &Request{
				ChannelID:   formVals.Get("channel_id"),
				ChannelName: formVals.Get("channel_name"),
				Command:     formVals.Get("command"),
				ResponseURL: formVals.Get("response_url"),
				TeamDomain:  formVals.Get("team_domain"),
				TeamID:      formVals.Get("team_id"),
				Text:        formVals.Get("text"),
				Token:       formVals.Get("token"),
				TriggerID:   formVals.Get("trigger_id"),
				UserID:      formVals.Get("user_id"),
				UserName:    formVals.Get("user_name"),
			}

			if server.handler != nil {
				msg, err := server.handler.Handle(request)

				if err != nil {
					ctx.Logger().Errorf("failed to handle request: %v", err)
					return ctx.JSON(http.StatusOK, slackError{
						ResponseType: "ephemeral",
						Text:         fmt.Sprintf("I can't handle request due to: %v", err),
					})
				}

				return ctx.JSON(http.StatusOK, msg)
			}

			return ctx.JSON(http.StatusOK, msgfmt.PlainText{
				Text: "I did my best.",
			})
		})


		errCh <- e.Start(fmt.Sprintf(":%d", server.listenPort))
	}()

	return errCh
}

func fromHeaderAsInt64(header http.Header, key string) int64 {
	valStr := header.Get(key)
	if valStr == "" {
		return 0
	}

	val, err := strconv.ParseInt(valStr, 10, 64)
	if err != nil {
		return 0
	}

	return val
}
