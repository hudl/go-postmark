// This package was modeled after Google's awesome go-github library
//   https://github.com/google/go-github
//
// This file contains a few helper functions from the project.

package postmark

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/google/go-querystring/query"
)

const (
	defaultBaseURL = "https://api.postmarkapp.com/"

	headerContentType  = "Content-Type"
	headerAccept       = "Accept"
	headerServerToken  = "X-Postmark-Server-Token"
	headerAccountToken = "X-Postmark-Account-Token"

	acceptType  = "application/json"
	contentType = "application/json"

	timeFormat = time.RFC3339
)

// A Client manages communication with the Postmark API.
type Client struct {
	// HTTP client used to communicate with the API.
	client *http.Client

	// BaseURL for the Postmark API.
	BaseURL *url.URL

	// Secret tokens used for authenticating with the Postmark API.
	ServerToken  string
	AccountToken string

	// Services used for talking to different parts of the Postmark API.
	Email *EmailService
}

// NewClient returns a new Postmark API client. If httpClient is nil,
// http.DefaultClient will be used.
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	baseURL, _ := url.Parse(defaultBaseURL)
	c := &Client{
		client:  httpClient,
		BaseURL: baseURL,
	}

	// configure services
	c.Email = &EmailService{client: c}

	return c
}

// addOptions adds the parameters in opt as URL query parameters to s.
// opt must be a struct whose fields may contain "url" tags.
//
// This function is from the google/go-github project
// https://github.com/google/go-github/blob/7277108aa3e8823e0e028f6c74aea2f4ce4a1b5a/github/github.go#L102-L122
func addOptions(s string, opt interface{}) (string, error) {
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return s, nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return s, err
	}

	qs, err := query.Values(opt)
	if err != nil {
		return s, err
	}

	u.RawQuery = qs.Encode()
	return u.String(), nil
}

// NewRequest creates an API request. A relative URL can be provided in path,
// in which case it is resolved relative to the BaseURL of the client.
// Relative URLs should always be specified without the preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body
func (c *Client) NewRequest(method, path string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(path)
	if err != nil {
		return nil, err
	}

	u := c.BaseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	err = CheckResponse(res)
	if err != nil {
		return res, err
	}

	if v != nil {
		err = json.NewDecoder(res.Body).Decode(v)
	}

	return res, err
}

// An ErrorResponse reports an error caused by an API request.
type ErrorResponse struct {
	Response  *http.Response
	ErrorCode int    `json:"ErrorCode"`
	Message   string `json:"Message"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("postmark: API error %d %q", e.ErrorCode, e.Message)
}

// CheckResponse checks the API response for errors, and returns them if
// present. A response is considered an error if it has a status code outside
// the 200 range.
func CheckResponse(r *http.Response) error {
	c := r.StatusCode
	if 200 <= c && c <= 299 {
		return nil
	}

	er := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		json.Unmarshal(data, er)
	}

	return er
}

// These helpers are from google's go-github
// https://github.com/google/go-github/blob/7277108aa3e8823e0e028f6c74aea2f4ce4a1b5a/github/github.go#L565-L588

// Bool is a helper function that allocates a new bool value to store v and
// returns a pointer to it.
func Bool(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

// Int is a helper function that allocates a new int value to store v and
// returns a pointer to it.
func Int(v int) *int {
	p := new(int)
	*p = v
	return p
}

// String is a helper function that allocates a new string value to store v and
// returns a pointer to it.
func String(v string) *string {
	p := new(string)
	*p = v
	return p
}
