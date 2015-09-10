package main

import (
	"fmt"

	"github.com/hudl/go-postmark/postmark"
)

func main() {
	client := postmark.NewClient(nil)
	client.ServerToken = "super-secret-server-token"

	email := &postmark.Email{
		To:       postmark.String("receiver@example.com"),
		From:     postmark.String("sender@example.com"),
		Subject:  postmark.String("Subject"),
		TextBody: postmark.String("Body"),
	}

	_, resp, err := client.Email.Send(email)
	if err != nil {
		fmt.Printf("Error sending email: %+v\n", err)
	} else {
		fmt.Printf("Email sent, postmark responded with: %+v", resp)
	}
}
