package protocol

import (
	"encoding/binary"
	"errors"
)

type ProtocolParser struct {
	messageLength uint16
	index         uint16
	command       []byte
	commandType   CommandType
}

func NewProtocolParser() *ProtocolParser {
	return &ProtocolParser{
		messageLength: 0,
		index:         0,
	}
}

// ParseStreamedData is to parse the byte data comes from server and convert it to meaningful internal commands
func (t *ProtocolParser) ParseStreamedData(readedByte byte) (interface{}, error) {

	// this is the first byte so we'll understand the command type here
	if t.index == 0 {
		t.commandType = CommandType(readedByte)
	}

	t.command = append(t.command, readedByte)
	t.index = t.index + 1

	if t.index > 2 {
		// 2nd and 3rd bytes are to store message length
		t.messageLength = binary.LittleEndian.Uint16(t.command[1:3])
	} else {
		return nil, nil
	}

	// if we are not complete yet, just add byte to array and return
	if t.index < t.messageLength {
		return nil, nil
	}

	if t.commandType == CommandTypeWhoAmI {
		// clear the array
		t.command = t.command[:0]
		t.index = 0
		if t.messageLength == 0 {
			return WhoAmICommand{}, nil
		}

		return WhoAmICommand{
			ClientID: binary.LittleEndian.Uint64(t.command[3:11]),
		}, nil
	} else if t.commandType == CommandTypeListClients {
		var clientIDs []uint64
		for i := 11; i <= int(t.messageLength); i = i + 8 {
			clientIDs = append(clientIDs, binary.LittleEndian.Uint64(t.command[i-8:i]))
		}
		// clear the array
		t.command = t.command[:0]
		t.index = 0
		return ListClientsCommand{
			ConnectedClients: clientIDs,
		}, nil
	} else if t.commandType == CommandTypeMessageFromClient {
		sender := binary.LittleEndian.Uint64(t.command[3:11])
		body := t.command[11:]

		// clear the array
		t.command = t.command[:0]
		t.index = 0
		return MessageFromClient{
			SenderID: sender,
			Body:     body,
		}, nil
	} else if t.commandType == CommandTypeSendMessage {
		// if message type is send message we store the recipient count in 4th and 5th bytes
		recipientsCount := binary.LittleEndian.Uint16(t.command[3:5])

		var recipients []uint64
		for i := 13; i < 13+int(recipientsCount*8); i = i + 8 {
			recipients = append(recipients, binary.LittleEndian.Uint64(t.command[i-8:i]))
		}
		messageBody := t.command[5+(recipientsCount*8):]
		// clear the array
		t.command = t.command[:0]
		t.index = 0
		return SendMessageCommand{
			Recipients: recipients,
			Body:       messageBody,
		}, nil
	} else {
		// clear the array
		t.command = t.command[:0]
		t.index = 0
		return nil, errors.New("Unknown Command Processed")
	}

}
