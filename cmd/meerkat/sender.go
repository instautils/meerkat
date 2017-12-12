package meerkat

type Sender interface {
	Send(to interface{}, message string) error
}

// TODO : user Sender interface for send messages to users.
// For example , implement sender interface for telegram.
