# go-postmark [![Build Status](https://travis-ci.org/hudl/go-postmark.svg?branch=master)](https://travis-ci.org/hudl/go-postmark)

`go-postmark` is a Go client library for interfacing with the
[Postmark API](http://developer.postmarkapp.com/), modeled after google's
awesome [google/go-github](http://github.com/google/go-github) library.

## Usage

Import the `postmark` package to get started.

```go
import "github.com/hudl/go-postmark/postmark"
```

Then construct a new Postmark client and set the Postmark service token.

```go
client := postmark.NewClient(nil)
client.ServiceToken = "super-secret-service-token"
```

You can use the various services registered with the client to access differnt
parts of the Postmark API. For exmaple, you can use the `Email` service to
interact with the [Email API](http://developer.postmarkapp.com/developer-api-email.html):

```go
resp, _, err := client.Email.Send(&postmark.Email{...})
resp, _, err := client.Email.SendBatch([]postmark.Email{...})
```

Check out more detailed examples in the [`examples`](./examples) directory.

### Helpers

The `Bool()`, `Int()` and `String()` helper functions in
[`postmark/postmark.go`](./postmark/postmark.go) are used to wrap values in
pointer variants for easier translation to JSON.

For example, to make an `Email` struct:
```go
email := &postmark.Email{
    To:         postmark.String("receiver@example.com"),
    From:       postmark.String("sender@example.com"),
    Subject:    postmark.String("Subject"),
    TextBody:   postmark.String("Body"),
    TrackOpens: postmark.Bool(true),
}
```

Pointers to values are used in many of the public types to show intent.
Meaning that if you pass a `nil` value to a struct field, it will be omitted.
Without knowing the intent of the creator, it would be impossible to
differentiate the zero-values for some of the primitive types, such as `int`
and `bool` from an intended zero-value. They would end up always be encoded to
JSON and sent to the Postmark API, possibly triggering API errors.

## Roadmap

This library is currently under development and has a limited subset of the
Postmark API implemented, specifically just the Email API. We plan to
eventually implement the entire Postmark API. Pull requests are welcome!

## License

This library is distributed under the MIT license found in the
[LICENSE](./LICENSE) file.
