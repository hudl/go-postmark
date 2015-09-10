package postmark_test

import (
	. "github.com/hudl/go-postmark/postmark"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

const (
	// The default base URL for the Postmark API.
	defaultBaseURL = "https://api.postmarkapp.com/"
)

// testEnv is a complete testing environment for mocking the Postmark API server.
type testEnv struct {
	// Mux is the HTTP request multiplexer used with the test server.
	Mux *http.ServeMux

	// Server is a test HTTP server used to provide mock API responses.
	Server *httptest.Server

	// Client is the Postmark client being tested
	Client *Client
}

// newTestEnv creates and configures a new test environment with a test HTTP
// server along with a Postmark client to talk to the server. Tests should
// register handlers on Mux which provide mock responses for the API method
// bing tested.
func newTestEnv() *testEnv {
	// test server
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	// Postmark client configured to use the test server
	client := NewClient(nil)
	url, _ := url.Parse(server.URL)
	client.BaseURL = url

	return &testEnv{
		Mux:    mux,
		Server: server,
		Client: client,
	}
}

// StopServer closes the test environment's HTTP server.
func (env *testEnv) StopServer() {
	env.Server.Close()
}

func TestPostmark(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Postmark Suite")
}
