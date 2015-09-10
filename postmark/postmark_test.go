package postmark_test

import (
	. "github.com/hudl/go-postmark/postmark"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

var _ = Describe("Postmark", func() {
	var (
		env    *testEnv
		client *Client
		err    error
		req    *http.Request
	)

	Describe("Creating a new client", func() {
		Context("without an http client", func() {
			BeforeEach(func() {
				client = NewClient(nil)
			})

			It("should use the default base URL", func() {
				Expect(client.BaseURL.String()).To(Equal(defaultBaseURL))
			})

			It("should register all services correctly", func() {
				Expect(client.Email).NotTo(BeNil())
			})
		})
	})

	Describe("Creating a new request", func() {
		BeforeEach(func() {
			client = NewClient(nil)
		})

		Context("with a valid http method, path and body", func() {
			// test type
			type T struct{ A int }

			BeforeEach(func() {
				req, err = client.NewRequest("GET", "/test", &T{A: 0})
			})

			It("should not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("should create a non-nil request", func() {
				Expect(req).NotTo(BeNil())
			})

			It("should expand the relative path", func() {
				testURL := fmt.Sprintf("%stest", defaultBaseURL)

				Expect(req.URL.String()).To(Equal(testURL))
			})

			It("should encode the body as valid JSON", func() {
				body, _ := ioutil.ReadAll(req.Body)

				Expect(body).To(MatchJSON(`{ "A": 0 }`))
			})
		})

		Context("with invalid JSON", func() {
			// test type with an unsupported json type
			type T struct{ A map[int]interface{} }

			It("should return a JSON unsupported type error", func() {
				_, err := client.NewRequest("GET", "/", &T{})
				Expect(err).NotTo(BeNil())

				_, ok := err.(*json.UnsupportedTypeError)
				Expect(ok).To(BeTrue())
			})
		})

		Context("with an invalid realtive path", func() {
			It("should return a URL parse error", func() {
				_, err := client.NewRequest("GET", ":", nil)
				Expect(err).NotTo(BeNil())

				urlErr, ok := err.(*url.Error)
				Expect(ok).To(BeTrue())
				Expect(urlErr.Op).To(Equal("parse"))
			})
		})

		Context("with an empty (nil) body", func() {
			BeforeEach(func() {
				req, err = client.NewRequest("GET", "/", nil)
			})

			It("should not return an error", func() {
				Expect(err).To(BeNil())
			})

			It("should have an empty (nil) body", func() {
				Expect(req.Body).To(BeNil())
			})
		})
	})

	Describe("Performing a request", func() {
		BeforeEach(func() {
			env = newTestEnv()
		})

		AfterEach(func() {
			env.StopServer()
		})

		Context("with a valid request", func() {
			It("should parse and return the correct response body", func() {
				// testing type
				type T struct{ A int }

				env.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					Expect(req.Method).To(Equal("GET"))
					fmt.Fprintf(w, `{ "A": 0 }`)
				})

				req, _ := env.Client.NewRequest("GET", "/", nil)
				body := new(T)
				env.Client.Do(req, body)

				Expect(body).To(Equal(&T{A: 0}))
			})
		})

		Context("when the response is an http error", func() {
			It("should return an error", func() {
				env.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					http.Error(w, "Bad Request", http.StatusBadRequest)
				})

				req, _ := env.Client.NewRequest("GET", "/", nil)
				_, err := env.Client.Do(req, nil)
				Expect(err).NotTo(BeNil())
			})
		})

		Context("when the native http client produces an error from a redirect loop", func() {
			It("should return a URL error", func() {
				env.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
					http.Redirect(w, r, "/", http.StatusFound)
				})

				req, _ := env.Client.NewRequest("GET", "/", nil)
				_, err := env.Client.Do(req, nil)
				Expect(err).NotTo(BeNil())

				_, ok := err.(*url.Error)
				Expect(ok).To(BeTrue())
			})
		})
	})

	Describe("Checking a response", func() {
		Context("with a valid response", func() {
			It("should return a Postmark error", func() {
				msg := `{
					"ErrorCode": 0,
					"Message": "Message"
				}`

				resp := &http.Response{
					Request:    &http.Request{},
					StatusCode: http.StatusBadRequest,
					Body:       ioutil.NopCloser(strings.NewReader(msg)),
				}

				err := CheckResponse(resp).(*ErrorResponse)
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(&ErrorResponse{
					Response:  resp,
					ErrorCode: 0,
					Message:   "Message",
				}))
			})
		})

		Context("with no body", func() {
			It("should return a Postmark error without an error code and message", func() {
				resp := &http.Response{
					Request:    &http.Request{},
					StatusCode: http.StatusBadRequest,
					Body:       ioutil.NopCloser(strings.NewReader("")),
				}

				err := CheckResponse(resp).(*ErrorResponse)
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(&ErrorResponse{
					Response: resp,
				}))
			})
		})
	})

	Describe("Stringifying a Postmark error response", func() {
		Context("with a non-nil Postmark error", func() {
			It("should return a non-empty string", func() {
				err := &ErrorResponse{}
				Expect(err.Error()).NotTo(BeEmpty())
			})
		})
	})
})
