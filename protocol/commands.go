package protocol

import (
	"encoding/binary"
	"errors"
)

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
	CommandLengthClient           = 8
)

// QueryCommand is used to send query to server
type QueryCommand struct {
}

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

// MessageFromClient is used for getting sent message to client
type MessageFromClient struct {
	SenderID uint64
	Body     []byte
}

// ToByteArray Converts WhoAmICommand to bytes
func (t *WhoAmICommand) ToByteArray() []byte {
	dataBytes := []byte{}

	clientIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(clientIDBytes, t.ClientID)
	dataBytes = append(dataBytes, clientIDBytes...)

	// calculate command length (commandType + messageLength + clientID)
	messageLength := CommandLengthType + CommandLengthMessageLength + 8

	// first byte is for command type
	command := []byte{uint8(CommandTypeWhoAmI)}

	// 2nd and 3rd bytes for messageLength
	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(messageLength))
	command = append(command, messageLengthBytes...)

	// rest is data
	if dataBytes != nil {
		command = append(command, dataBytes...)
	}
	return command
}

// ToByteArray Converts ListClientsCommand to bytes
func (t *ListClientsCommand) ToByteArray(clientID uint64) []byte {
	dataBytes := []byte{}

	connectedClientBytes := make([]byte, 8)
	for i := 0; i < len(t.ConnectedClients); i++ {
		if t.ConnectedClients[i] != clientID {
			binary.LittleEndian.PutUint64(connectedClientBytes, t.ConnectedClients[i])
			dataBytes = append(dataBytes, connectedClientBytes...)
		}
	}

	// calculate command length (commandType + messageLength + (clientCount-1)*8)
	messageLength := CommandLengthType + CommandLengthMessageLength + (8 * (len(t.ConnectedClients) - 1))
	// first byte is for command type
	command := []byte{uint8(CommandTypeListClients)}

	// 2nd and 3rd bytes for messageLength
	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(messageLength))
	command = append(command, messageLengthBytes...)

	// rest is data
	if dataBytes != nil {
		command = append(command, dataBytes...)
	}
	return command
}

// ToByteArray Converts MessageFromClient to bytes
func (t *MessageFromClient) ToByteArray() []byte {
	dataBytes := []byte{}

	senderBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(senderBytes, t.SenderID)

	bodyBytes := t.Body

	dataBytes = append(dataBytes, senderBytes...)
	dataBytes = append(dataBytes, bodyBytes...)

	// calculate command length (commandType + messageLength + senderid + messagebody)
	messageLength := CommandLengthType + CommandLengthMessageLength + CommandLengthClient + len(bodyBytes)
	// first byte is for command type
	command := []byte{uint8(CommandTypeMessageFromClient)}

	// 2nd and 3rd bytes for messageLength
	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(messageLength))
	command = append(command, messageLengthBytes...)

	// rest is data
	if dataBytes != nil {
		command = append(command, dataBytes...)
	}
	return command
}

// ToByteArray Converts SendMessageCommand to bytes
func (t *SendMessageCommand) ToByteArray() []byte {

	dataBytes := []byte{}

	// we have a special case here, 4th and 5th bytes for recipient length, send them as data
	recipientLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(recipientLengthBytes, uint16(len(t.Recipients)))
	dataBytes = append(dataBytes, recipientLengthBytes...)

	// add recipients to the data
	recipientIDBytes := make([]byte, 8)
	for i := 0; i < len(t.Recipients); i++ {
		binary.LittleEndian.PutUint64(recipientIDBytes, t.Recipients[i])
		dataBytes = append(dataBytes, recipientIDBytes...)
	}

	// and finally add the message body
	dataBytes = append(dataBytes, t.Body...)

	// calculate message length
	// commandType + messageLength + recipientsLength + recipients + messageBody
	messageLength := CommandLengthType + CommandLengthMessageLength + CommandLengthRecipientsLength + (len(t.Recipients) * 8) + len(t.Body)
	// first byte is for command type
	command := []byte{uint8(CommandTypeSendMessage)}

	// 2nd and 3rd bytes for messageLength
	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(messageLength))
	command = append(command, messageLengthBytes...)

	// rest is data
	if dataBytes != nil {
		command = append(command, dataBytes...)
	}
	return command
}

// CreateQueryCommand Creates a query command for client to send to server
func (t *QueryCommand) CreateQueryCommand(commandType CommandType) []byte {
	// first byte is for command type
	command := []byte{uint8(commandType)}
	// 2nd and 3rd bytes for messageLength
	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(0))
	command = append(command, messageLengthBytes...)
	return command

}
