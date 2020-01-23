package protocol

import "errors"

var (
	UnknownCommand = errors.New("Unknown command")
)

// WhoAmICommand is used for getting client id
type WhoAmICommand struct {
	ClientID uint64
}

// ListClientsCommand is used for getting all connected clients
type ListClientsCommand struct {
	ConnectedClients []uint64
}

// SendMessageCommand is used for sending message to selected clients
type SendMessageCommand struct {
	Recipients []uint64
	Body       []byte
}
