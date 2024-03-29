package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
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
	handlers      []handler
	middlewares   []echo.MiddlewareFunc
}

type handler func() (method, path string, handlerFunc echo.HandlerFunc, isForSlack bool)

type Option func(*Server) error

func New(signingSecret string, opts ...Option) (*Server, error) {
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

func Handlers(handlers ...handler) Option {
	return func(server *Server) error {
		server.handlers = handlers

		return nil
	}
}

func Middlewares(middlewares ...echo.MiddlewareFunc) Option {
	return func(server *Server) error {
		server.middlewares = middlewares

		return nil
	}
}

func (server *Server) StartServer() <-chan error {
	// make error chan
	errCh := make(chan error, 1)

	go func() {
		// make echo
		e := echo.New()
		e.HideBanner = true
		e.HidePort = true

		e.Use(middleware.Logger())
		e.Logger.SetLevel(server.logLevel)

		if server.signingSecret == "" {
			errCh <- errors.New("empty signing secret")
			return
		}

		// make verifier
		verifier := NewVerifier(server.signingSecret)

		// register other middlewares
		if len(server.middlewares) > 1 {
			e.Use(server.middlewares...)
		}

		// register handlers
		for _, h := range server.handlers {
			method, path, handlerFunc, isForSlack := h()

			var handlerMiddlewares []echo.MiddlewareFunc
			if isForSlack {
				handlerMiddlewares = []echo.MiddlewareFunc{verifier.Middleware()}
			}

			switch method {
			case http.MethodPost:
				e.POST(path, handlerFunc, handlerMiddlewares...)
			case http.MethodGet:
				e.GET(path, handlerFunc, handlerMiddlewares...)
			case http.MethodPut:
				e.PUT(path, handlerFunc, handlerMiddlewares...)
			case http.MethodHead:
				e.HEAD(path, handlerFunc, handlerMiddlewares...)
			case http.MethodConnect:
				e.CONNECT(path, handlerFunc, handlerMiddlewares...)
			case http.MethodDelete:
				e.DELETE(path, handlerFunc, handlerMiddlewares...)
			case http.MethodOptions:
				e.OPTIONS(path, handlerFunc, handlerMiddlewares...)
			case http.MethodPatch:
				e.PATCH(path, handlerFunc, handlerMiddlewares...)
			case http.MethodTrace:
				e.TRACE(path, handlerFunc, handlerMiddlewares...)
			default:
				errCh <- fmt.Errorf("unknwon method '%s'", method)
				return
			}
		}

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
