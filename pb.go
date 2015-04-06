// Copyright 2015 Demisto. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// pb package implements pandorabots API as seen in https://developer.pandorabots.com/docs
package pb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultURL = "https://aiaas.pandorabots.com"
	bot        = "bot"
	talk       = "talk"
)

var (
	// The error raised when credentials are not provided
	ErrNoCred = errors.New("Missing application ID or user key")
)

// Client interacts with the services provided by pandorabots.
type Client struct {
	appId    string       // ID of the application we are using
	userKey  string       // The user credentials to access the API
	url      string       // The URL for the API.
	errorlog *log.Logger  // Optional logger to write errors to
	tracelog *log.Logger  // Optional logger to write trace and debug data to
	c        *http.Client // The client to use for requests
}

// OptionFunc is a function that configures a Client.
// It is used in New
type OptionFunc func(*Client) error

// errorf logs to the error log.
func (c *Client) errorf(format string, args ...interface{}) {
	if c.errorlog != nil {
		c.errorlog.Printf(format, args...)
	}
}

// tracef logs to the trace log.
func (c *Client) tracef(format string, args ...interface{}) {
	if c.tracelog != nil {
		c.tracelog.Printf(format, args...)
	}
}

// New creates a new pandorabots client.
//
// The caller can configure the new client by passing configuration options to the func.
//
// Example:
//
//   client, err := pb.New(
//     pb.SetErrorLog(log.New(os.Stderr, "PB: ", log.Lshortfile),
//     pb.SetCredentials(appId, userKey))
//
// If no URL is configured, Client uses DefaultURL by default.
//
// If no HttpClient is configured, then http.DefaultClient is used.
// You can use your own http.Client with some http.Transport for advanced scenarios.
//
// An error is also returned when some configuration option is invalid.
func New(options ...OptionFunc) (*Client, error) {
	// Set up the client
	c := &Client{
		c: http.DefaultClient,
	}

	// Run the options on it
	for _, option := range options {
		if err := option(c); err != nil {
			return nil, err
		}
	}
	if c.appId == "" || c.userKey == "" {
		c.errorf(ErrNoCred.Error())
		return nil, ErrNoCred
	}
	if c.url == "" {
		c.url = DefaultURL
	}
	if strings.HasSuffix(c.url, "/") {
		c.url = c.url[0 : len(c.url)-1]
	}
	c.tracef("Using URL [%s]\n", c.url)

	return c, nil
}

// Initialization functions

// SetCredentials sets the app id and user key to use with pandorabots
func SetCredentials(appId, userKey string) OptionFunc {
	return func(c *Client) error {
		c.appId, c.userKey = appId, userKey
		return nil
	}
}

// SetHttpClient can be used to specify the http.Client to use when making
// HTTP requests to pandorabots.
func SetHttpClient(httpClient *http.Client) OptionFunc {
	return func(c *Client) error {
		if httpClient != nil {
			c.c = httpClient
		} else {
			c.c = http.DefaultClient
		}
		return nil
	}
}

// SetUrl defines the URL endpoint for pandorabots
func SetUrl(rawurl string) OptionFunc {
	return func(c *Client) error {
		if rawurl == "" {
			rawurl = DefaultURL
		}
		u, err := url.Parse(rawurl)
		if err != nil {
			c.errorf("Invalid URL [%s] - %v\n", rawurl, err)
			return err
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			msg := fmt.Sprintf("Invalid schema specified [%s]", rawurl)
			c.errorf(msg)
			return errors.New(msg)
		}
		c.url = rawurl
		return nil
	}
}

// SetErrorLog sets the logger for critical messages. It is nil by default.
func SetErrorLog(logger *log.Logger) func(*Client) error {
	return func(c *Client) error {
		c.errorlog = logger
		return nil
	}
}

// SetTraceLog specifies the logger to use for output of trace messages like
// HTTP requests and responses. It is nil by default.
func SetTraceLog(logger *log.Logger) func(*Client) error {
	return func(c *Client) error {
		c.tracelog = logger
		return nil
	}
}

// dumpRequest dumps a request to the debug logger if it was defined
func (c *Client) dumpRequest(req *http.Request) {
	if c.tracelog != nil {
		out, err := httputil.DumpRequestOut(req, false)
		if err == nil {
			c.tracef("%s\n", string(out))
		}
	}
}

// dumpResponse dumps a response to the debug logger if it was defined
func (c *Client) dumpResponse(resp *http.Response) {
	if c.tracelog != nil {
		out, err := httputil.DumpResponse(resp, true)
		if err == nil {
			c.tracef("%s\n", string(out))
		}
	}
}

// Request handling functions

// handleError will handle responses with status code different from success
func (c *Client) handleError(resp *http.Response) error {
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if c.errorlog != nil {
			out, err := httputil.DumpResponse(resp, true)
			if err == nil {
				c.errorf("%s\n", string(out))
			}
		}
		msg := fmt.Sprintf("Unexpected status code: %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
		c.errorf(msg)
		return errors.New(msg)
	}
	return nil
}

func (c *Client) appUrl(action string) string {
	return fmt.Sprintf("%s/%s/%s", c.url, action, c.appId)
}

func (c *Client) botUrl(action, botName string) string {
	return c.appUrl(action) + "/" + botName
}

// do executes the API request.
// Returns the response if the status code is between 200 and 299
// `body` is an optional body for the POST requests.
func (c *Client) do(method, rawurl string, params map[string]string, body io.Reader, result interface{}) error {
	values := url.Values{}
	values.Set("user_key", c.userKey)
	for k, v := range params {
		values.Add(k, v)
	}

	req, err := http.NewRequest(method, rawurl+"?"+values.Encode(), body)
	if err != nil {
		return err
	}
	c.dumpRequest(req)

	resp, err := c.c.Do(req)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if err = c.handleError(resp); err != nil {
		return err
	}
	c.dumpResponse(resp)
	if result != nil {
		switch result.(type) {
		// Should we just dump the response body
		case io.Writer:
			if _, err = io.Copy(result.(io.Writer), resp.Body); err != nil {
				return err
			}
		default:
			if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
				return err
			}
		}
	}
	return nil
}

type BotEntry struct {
	Name        string `json:"botname"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Compiled    string `json:"compiled"`
	Open        string `json:"open"`
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/listBots
func (c *Client) List() ([]BotEntry, error) {
	result := make([]BotEntry, 0)
	err := c.do("GET", c.appUrl(bot), nil, nil, &result)
	return result, err
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/createBot
func (c *Client) CreateBot(name string) error {
	return c.do("PUT", c.botUrl(bot, name), nil, nil, nil)
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/deleteBot
func (c *Client) DeleteBot(name string) error {
	return c.do("DELETE", c.botUrl(bot, name), nil, nil, nil)
}

type BotFile struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	Modified  time.Time `json:"modified"`
	LoadOrder int       `json:"loadorder"`
	Items     int       `json:"items"`
}

type BotFiles struct {
	Username      string    `json:"username"`
	Appname       string    `json:"appname"`
	Botname       string    `json:"botname"`
	Description   string    `json:"description"`
	Language      string    `json:"language"`
	Created       time.Time `json:"created"`
	Open          string    `json:"open"`
	Files         []BotFile `json:"files"`
	Sets          []BotFile `json:"sets"`
	Maps          []BotFile `json:"maps"`
	Substitutions []BotFile `json:"substitutions"`
	Properties    []BotFile `json:"properties"`
	Pdefaults     []BotFile `json:"pdefaults"`
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/listBotFiles
func (c *Client) ListFiles(name string) (BotFiles, error) {
	var result BotFiles
	err := c.do("GET", c.botUrl(bot, name), nil, nil, &result)
	return result, err
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/listBotFiles
func (c *Client) DownloadFiles(name string, zip io.Writer) error {
	return c.do("GET", c.botUrl(bot, name), map[string]string{"return": "zip"}, nil, zip)
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/listBotFiles
func (c *Client) DownloadFilesToPath(name, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.do("GET", c.botUrl(bot, name), map[string]string{"return": "zip"}, nil, f)
}

func (c *Client) fileToUrl(name, filename string) (string, error) {
	rawurl := c.botUrl(bot, name)
	ext := filepath.Ext(filename)
	switch ext {
	case ".aiml":
		rawurl += "/file/" + filename
	case ".set", ".map", ".substitution":
		rawurl += "/" + ext[1:] + "/" + filename[0:len(filename)-len(ext)]
	case ".properties", ".pdefaults":
		rawurl += "/" + ext[1:]
	default:
		return "", fmt.Errorf("Extension is not recognized [%s]", ext)
	}
	return rawurl, nil
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/uploadFile1
// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/uploadFile2
func (c *Client) UploadFile(name, filename string, data io.Reader) error {
	rawurl, err := c.fileToUrl(name, filename)
	if err != nil {
		return err
	}
	return c.do("PUT", rawurl, nil, data, nil)
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/uploadFile1
// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/uploadFile2
func (c *Client) UploadFileFromPath(name, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	rawurl, err := c.fileToUrl(name, filepath.Base(path))
	if err != nil {
		return err
	}
	return c.do("PUT", rawurl, nil, f, nil)
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/deleteBotFile1
// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/deleteBotFile2
func (c *Client) DeleteFile(name, filename string) error {
	rawurl, err := c.fileToUrl(name, filename)
	if err != nil {
		return err
	}
	return c.do("DELETE", rawurl, nil, nil, nil)
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/getBotFile1
// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/getBotFile2
func (c *Client) GetFile(name, filename string, out io.Writer) error {
	rawurl, err := c.fileToUrl(name, filename)
	if err != nil {
		return err
	}
	return c.do("GET", rawurl, nil, nil, out)
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/getBotFile1
// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/getBotFile2
func (c *Client) GetFileToPath(name, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	rawurl, err := c.fileToUrl(name, filepath.Base(path))
	if err != nil {
		return err
	}
	return c.do("GET", rawurl, nil, nil, f)
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/compileBot
func (c *Client) Verify(name string) error {
	return c.do("GET", c.botUrl(bot, name)+"/verify", nil, nil, nil)
}

type Reply struct {
	SessionId int      `json:"sessionid"`
	Responses []string `json:"responses"`
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/talkBot
func (c *Client) Talk(name, input, clientName string, sessionId int, recent bool) (*Reply, error) {
	return c.TalkDebug(name, input, clientName, sessionId, recent, "", "", false, false, false, false)
}

// See https://developer.pandorabots.com/docs#!/pandorabots_api_swagger_1_2_beta/debugBot
func (c *Client) TalkDebug(name, input, clientName string, sessionId int, recent bool, that, topic string, extra, reset, trace, reload bool) (*Reply, error) {
	params := make(map[string]string)
	params["input"] = input
	if clientName != "" {
		params["client_name"] = clientName
	}
	if sessionId != 0 {
		params["sessionid"] = strconv.Itoa(sessionId)
	}
	if recent {
		params["recent"] = "true"
	}
	if that != "" {
		params["that"] = that
	}
	if topic != "" {
		params["topic"] = topic
	}
	if extra {
		params["extra"] = "true"
	}
	if reset {
		params["reset"] = "true"
	}
	if trace {
		params["trace"] = "true"
	}
	if reload {
		params["reload"] = "true"
	}
	var reply Reply
	err := c.do("POST", c.botUrl(talk, name), params, nil, &reply)
	return &reply, err
}
