package protocol

import "errors"

var (
	UnknownCommand = errors.New("Unknown command")
)

// CommandType is an enumator for command types
type CommandType uint8

const (
	// CommandTypeWhoAmI Command
	CommandTypeWhoAmI CommandType = 1
	// CommandTypeListClients Command
	CommandTypeListClients CommandType = 2
	// CommandTypeSendMessage Command
	CommandTypeSendMessage CommandType = 3
	// CommandTypeMessageFromClient Command
	CommandTypeMessageFromClient CommandType = 4
	// CommandTypeUnknown Command
	CommandTypeUnknown CommandType = 0
)

const (
	CommandLengthType             = 1
	CommandLengthMessageLength    = 2
	CommandLengthRecipientsLength = 2
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

type MessageFromClient struct {
	SenderID uint64
	Body     []byte
}
