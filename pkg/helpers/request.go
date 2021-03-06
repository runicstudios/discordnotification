package helpers

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
	"uwdiscorwb/v1/pkg/log"
)

var mutex = sync.Mutex{}

// Client provides a wrapper around http client
type Client struct {
	Client   *http.Client
	Headers  http.Header
	BasePath string
}

type Response struct {
	Res *http.Response
}

type Request struct {
	Req    *http.Request
	Client *http.Client
}

// global client options
type GlobalOptions struct {
	Timeout  time.Duration
	BasePath string
	Headers  http.Header
}

// options object
type Options struct {
	Url     string
	Method  string
	Headers http.Header
	Body    interface{}
	Query   map[string]string
}

var client *Client

// create a helpers client with global configurations
func NewClient() *Client {
	if client != nil {
		return client
	}

	// acquiring lock for creating a singleton client of helpers library
	mutex.Lock()
	defer mutex.Unlock()

	client = &Client{}
	// creating default transport logic for helpers
	// default to 30 seconds timeout for any outgoing requests
	var transport = http.DefaultTransport
	transport.(*http.Transport).TLSClientConfig = &tls.Config{}
	transport.(*http.Transport).ForceAttemptHTTP2 = false
	client.Client = &http.Client{
		Transport:     transport,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       30 * time.Second,
	}
	headers := http.Header{}
	headers.Add("Content-Type", "application/json")
	client.Headers = headers
	return client
}

// response wrapper for returning bytes of data returned by response
func (wr *Response) GetBody() ([]byte, error) {
	defer wr.Res.Body.Close()
	return ioutil.ReadAll(wr.Res.Body)
}

// response wrapper for returning headers filed from response
func (wr *Response) GetHeaders() http.Header {
	return wr.Res.Header
}

// response wrapper for returning the status code
func (wr *Response) GetStatusCode() int {
	return wr.Res.StatusCode
}

// creates a new helpers object, this needs to be invoke for every call
// client will hold all the global info about the helpers
func (client *Client) NewRequest(options Options) (*Request, error) {
	var err error
	uri := uriBuilder(client.BasePath, options.Url, options.Query)
	var req *http.Request
	if options.Body == nil {
		req, err = http.NewRequest(options.Method, uri, nil)
		if err != nil {
			return nil, err
		}
	} else if _, ok := options.Body.([]byte); ok {
		reader := bytes.NewReader(options.Body.([]byte))
		req, err = http.NewRequest(options.Method, uri, reader)
		if err != nil {
			return nil, err
		}
	} else {
		body, err := requestBodyBuilder(options.Body)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(options.Method, uri, body)
		if err != nil {
			return nil, err
		}
	}
	req.Header = client.Headers
	if len(options.Headers) != 0 {
		req.Header = options.Headers
	}
	return &Request{
		Req:    req,
		Client: client.Client,
	}, nil
}

func (client *Client) AddHeader(key, value string) http.Header {
	client.Headers.Set(key, value)
	return client.Headers
}

// converts the body to json, if body is empty returns null or empty
// returns a buffer of stringifies struct or helpers body
func requestBodyBuilder(body interface{}) (*bytes.Buffer, error) {
	if body == nil {
		return nil, nil
	}
	var reader *bytes.Buffer
	switch body.(type) {
	case string:
		log.Debug("converting the helpers body to bytes of string")
		reader = bytes.NewBuffer([]byte(body.(string)))
	default:
		log.Debug("trying to convert the helpers body to json")
		mr, err := json.Marshal(body)
		if err != nil {
			log.Debug(fmt.Sprintf("failed json marshall - %v", err))
			return nil, err
		}
		log.Debug(fmt.Sprintf("stringifies json helpers body - %v", string(mr)))
		reader = bytes.NewBuffer(mr)
	}
	return reader, nil
}

// triggers the api call and returns error or response object
func (req *Request) Send() (*Response, error) {
	var wr Response
	// calling the api
	resp, err := req.Client.Do(req.Req)
	// pass the api response to current client wrapper
	wr.Res = resp
	return &wr, err
}

// builds the uri with all the query parameters and if basepath is provided attaches that as well
// if basepath and full url both provided, full url will take precedence
func uriBuilder(basePath string, urlPath string, qp map[string]string) string {
	qString := ""
	for k, v := range qp {
		qString += k + "=" + v
	}
	if strings.Contains(urlPath, "?") {
		urlPath += qString
	}
	if qString != "" {
		urlPath += "?" + qString
	}
	if strings.HasPrefix(urlPath, "http") {
		urlPath += qString
	} else if strings.HasSuffix(basePath, "/") {
		urlPath = basePath + urlPath
	} else {
		urlPath = basePath + urlPath
	}
	log.Debug("generated url: ", urlPath)
	return urlPath
}
