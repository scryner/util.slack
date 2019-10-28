package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/scryner/util.slack/internal/lrucache"
)

const (
	defaultServerAddr     = "https://slack.com"
	defaultLruCacheCapacity  = 2048
	defaultRequestTimeout = time.Second * 10
)

type API struct {
	serverAddr     string
	botAccessToken string
	requestTimeout time.Duration
	cacheCapacity  int

	httpCli          *http.Client
	emailToUserCache Cache
	idToUserCache    Cache
}

type Option func(*API) error

func ServerAddress(serverAddr string) Option {
	return func(api *API) error {
		api.serverAddr = serverAddr
		return nil
	}
}

func RequestTimeout(timeout time.Duration) Option {
	return func(api *API) error {
		api.requestTimeout = timeout
		return nil
	}
}

func EmailToUserCache(cache Cache) Option {
	return func(api *API) error {
		api.emailToUserCache = cache
		return nil
	}
}

func IdToUserCache(cache Cache) Option {
	return func(api *API) error {
		api.idToUserCache = cache
		return nil
	}
}

func CacheCapacity(capacity int) Option {
	return func(api *API) error {
		api.cacheCapacity = capacity
		return nil
	}
}

func NewAPI(botAccessToken string, opts ...Option) (*API, error) {
	api := &API{
		serverAddr:     defaultServerAddr,
		botAccessToken: botAccessToken,
		requestTimeout: defaultRequestTimeout,
	}

	var err error
	for _, opt := range opts {
		err = opt(api)
		if err != nil {
			return nil, err
		}
	}

	api.httpCli = &http.Client{
		Timeout: api.requestTimeout,
	}

	if api.emailToUserCache == nil {
		api.emailToUserCache = lrucache.NewCache(defaultLruCacheCapacity)
	}

	if api.idToUserCache == nil {
		api.idToUserCache = lrucache.NewCache(defaultLruCacheCapacity)
	}

	return api, nil
}

func (api *API) doHTTPGet(apiPath string, params url.Values) (*http.Response, error) {
	// clone params
	newParams := make(url.Values)

	for key, val := range params {
		newParams[key] = val
	}

	// adding bot access token
	newParams.Set("token", api.botAccessToken)

	u := fmt.Sprintf("%s/%s?%s", api.serverAddr, apiPath, newParams.Encode())

	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make http request: %v", err)
	}

	return api.httpCli.Do(req)
}

func (api *API) doHTTPPost(apiPath string, params url.Values) (*http.Response, error) {
	// clone params
	newParams := make(url.Values)

	for key, val := range params {
		newParams[key] = val
	}

	// adding bot access token
	newParams.Set("token", api.botAccessToken)

	u := fmt.Sprintf("%s/%s?%s", api.serverAddr, apiPath, newParams.Encode())

	req, err := http.NewRequest(http.MethodPost, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to make http request: %v", err)
	}

	return api.httpCli.Do(req)
}

func (api *API) doHTTPPostJSON(apiPath string, params url.Values, v interface{}) (*http.Response, error) {
	// marshal content
	content, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content to JSON: %v", err)
	}

	var u string
	if params == nil {
		u = fmt.Sprintf("%s/%s", api.serverAddr, apiPath)
	} else {
		u = fmt.Sprintf("%s/%s?%s", api.serverAddr, apiPath, params.Encode())
	}

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to make http request: %v", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", api.botAccessToken))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	return api.httpCli.Do(req)
}
