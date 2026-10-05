package email

import "context"

type Address struct {
	Name  string
	Email string
}

type Message struct {
	To       Address
	Subject  string
	HTMLBody string
	TextBody string
}

type Sender interface {
	Send(ctx context.Context, message Message) error
}
