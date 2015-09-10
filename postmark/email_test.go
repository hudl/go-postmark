package postmark_test

import (
	. "github.com/hudl/go-postmark/postmark"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

// encodeBase64 is a helper function that base64 encodes a string and returns
// a string
func encodeBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// decodeBase64 is a helper function that decodes a base64 encoded string and
// returns a string or error.
func decodeBase64(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	return string(data), err
}

var _ = Describe("Email", func() {
	var env *testEnv

	assertEmailHeaders := func(r *http.Request) {
		Expect(r.Header.Get("Content-Type")).To(Equal("application/json"))
		Expect(r.Header.Get("Accept")).To(Equal("application/json"))
		Expect(r.Header.Get("X-Postmark-Server-Token")).To(Equal(env.Client.ServerToken))
	}

	Describe("Marshaling an attachment", func() {
		var (
			content        string
			contentBase64  string
			attachment     *Attachment
			attachmentJSON string
		)

		BeforeEach(func() {
			content = "Content"
			contentBase64 = encodeBase64(content)
			attachment = &Attachment{
				Name:        String("Name"),
				Content:     String(content),
				ContentType: String("ContentType"),
				ContentID:   String("ContentID"),
			}
			attachmentJSON = fmt.Sprintf(`{
				"Name": "Name",
				"Content": "%s",
				"ContentType": "ContentType",
				"ContentID": "ContentID"
			}`, contentBase64)
		})

		Context("to JSON", func() {
			It("should not return an error", func() {
				_, err := json.Marshal(attachment)
				Expect(err).To(BeNil())
			})

			It("should base64 encode the content", func() {
				a, _ := json.Marshal(attachment)
				Expect(a).To(MatchJSON(attachmentJSON))
			})
		})

		Context("from JSON", func() {
			It("should not return an error", func() {
				a := new(Attachment)
				err := json.Unmarshal([]byte(attachmentJSON), a)
				Expect(err).To(BeNil())
			})

			It("should base64 decode the content", func() {
				a := new(Attachment)
				json.Unmarshal([]byte(attachmentJSON), a)
				Expect(*a.Content).To(Equal(content))
			})
		})
	})

	Describe("Sending an email", func() {
		BeforeEach(func() {
			env = newTestEnv()
		})

		AfterEach(func() {
			env.StopServer()
		})

		Context("with a valid email", func() {
			BeforeEach(func() {
				env.Mux.HandleFunc("/email", func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, `{
						"To": "receiver@example.com",
						"MessageID": "MessageID"
					}`)
				})
			})

			It("should not return an error", func() {
				_, _, err := env.Client.Email.Send(&Email{})
				Expect(err).To(BeNil())
			})

			It("should post to the /email endpoint", func() {
				_, resp, _ := env.Client.Email.Send(&Email{})
				Expect(resp.Request.Method).To(Equal("POST"))
				Expect(resp.Request.URL.Path).To(Equal("/email"))
			})

			It("should use the correct headers", func() {
				_, resp, _ := env.Client.Email.Send(&Email{})
				assertEmailHeaders(resp.Request)
			})

			It("should return the correct result", func() {
				result, _, _ := env.Client.Email.Send(&Email{})
				Expect(result).To(Equal(&EmailResult{
					To:        "receiver@example.com",
					MessageID: "MessageID",
				}))
			})
		})

		Context("when an API error is returned", func() {
			BeforeEach(func() {
				env.Mux.HandleFunc("/email", func(w http.ResponseWriter, r *http.Request) {
					// TODO: create constants for http errors defined by Postmark API
					w.WriteHeader(422) // Unprocessable Entity
					fmt.Fprintf(w, `{
						"ErrorCode": 0,
						"Message": "Message"
					}`)
				})
			})

			It("should return a Postmark error", func() {
				_, resp, err := env.Client.Email.Send(&Email{})
				Expect(err).NotTo(BeNil())
				Expect(err).To(Equal(&ErrorResponse{
					Response:  resp,
					ErrorCode: 0,
					Message:   "Message",
				}))
			})
		})
	})

	Describe("Sending a batch email", func() {
		BeforeEach(func() {
			env = newTestEnv()
		})

		AfterEach(func() {
			env.StopServer()
		})

		Context("with a valid batch email", func() {
			BeforeEach(func() {
				env.Mux.HandleFunc("/email/batch", func(w http.ResponseWriter, r *http.Request) {
					fmt.Fprintf(w, `[{
						"To": "receiver1@example.com",
						"MessageID": "MessageID1"
					}, {
						"To": "receiver2@example.com",
						"MessageID": "MessageID2"
					}]`)
				})
			})

			It("should not return an error", func() {
				_, _, err := env.Client.Email.SendBatch([]Email{})
				Expect(err).To(BeNil())
			})

			It("should post to the /email endpoint", func() {
				_, resp, _ := env.Client.Email.SendBatch([]Email{})
				Expect(resp.Request.Method).To(Equal("POST"))
				Expect(resp.Request.URL.Path).To(Equal("/email/batch"))
			})

			It("should use the correct headers", func() {
				_, resp, _ := env.Client.Email.SendBatch([]Email{})
				assertEmailHeaders(resp.Request)
			})

			It("should return the correct results", func() {
				results, _, _ := env.Client.Email.SendBatch([]Email{})
				Expect(results).To(Equal([]EmailResult{
					EmailResult{
						To:        "receiver1@example.com",
						MessageID: "MessageID1",
					},
					EmailResult{
						To:        "receiver2@example.com",
						MessageID: "MessageID2",
					},
				}))
			})
		})
	})
})
