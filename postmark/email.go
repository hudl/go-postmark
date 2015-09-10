package postmark

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
)

// EmailService handles communication with the Email related
// methods of the Postmark API.
type EmailService struct {
	client *Client
}

type Email struct {
	From        *string      `json:"From,omitempty"`
	To          *string      `json:"To,omitempty"`
	Cc          *string      `json:"Cc,omitempty"`
	Bcc         *string      `json:"Bcc,omitempty"`
	Subject     *string      `json:"Subject,omitempty"`
	Tag         *string      `json:"Tag,omitempty"`
	HTMLBody    *string      `json:"HtmlBody,omitempty"`
	TextBody    *string      `json:"TextBody,omitempty"`
	ReplyTo     *string      `json:"ReplyTo,omitempty"`
	Headers     []Header     `json:"Headers,omitempty"`
	TrackOpens  *bool        `json:"TrackOpens,omitempty"`
	Attachments []Attachment `json:"Attachments,omitempty"`
}

type Header struct {
	Name  *string `json:"Name,omitempty"`
	Value *string `json:"Value,omitempty"`
}

type Attachment struct {
	Name        *string `json:"Name,omitempty"`
	Content     *string `json:"Content,omitempty"`
	ContentType *string `json:"ContentType,omitempty"`
	ContentID   *string `json:"ContentID,omitempty"`
}

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

// type alias used to avoid recursive definition in MarshalJSON and
// UnmarshalJSON
type attachment Attachment

func (a *Attachment) MarshalJSON() ([]byte, error) {
	temp := attachment(*a)

	// encode the content using base64 as specified by the Postmark API
	*temp.Content = encodeBase64(*a.Content)

	return json.Marshal(temp)
}

func (a *Attachment) UnmarshalJSON(b []byte) error {
	temp := attachment{}
	err := json.Unmarshal(b, &temp)
	if err != nil {
		return err
	}

	*a = Attachment(temp)
	*a.Content, _ = decodeBase64(*temp.Content)
	return nil
}

type EmailResult struct {
	To          string
	SubmittedAt *time.Time
	MessageID   string
}

func (s *EmailService) Send(email *Email) (*EmailResult, *http.Response, error) {
	req, err := s.client.NewRequest("POST", "email", email)
	if err != nil {
		return nil, nil, err
	}

	// set headers
	req.Header.Set(headerContentType, contentType)
	req.Header.Set(headerAccept, acceptType)
	req.Header.Set(headerServerToken, s.client.ServerToken)

	result := new(EmailResult)
	resp, err := s.client.Do(req, result)
	if err != nil {
		return nil, resp, err
	}

	return result, resp, err
}

func (s *EmailService) SendBatch(emails []Email) ([]EmailResult, *http.Response, error) {
	req, err := s.client.NewRequest("POST", "email/batch", emails)
	if err != nil {
		return nil, nil, err
	}

	// set headers
	req.Header.Set(headerContentType, contentType)
	req.Header.Set(headerAccept, acceptType)
	req.Header.Set(headerServerToken, s.client.ServerToken)

	results := new([]EmailResult)
	resp, err := s.client.Do(req, results)
	if err != nil {
		return nil, resp, err
	}

	return *results, resp, err
}
