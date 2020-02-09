package protocol

import (
	"encoding/binary"
	"errors"
)

const (
	index_CommandTypeStart      = 0
	index_CommandTypeEnd        = 1
	index_MessageLengthStart    = 1
	index_MessageLengthEnd      = 3
	index_RecipientsLengthStart = 3
	index_RecipientsLengthEnd   = 5
)

type ProtocolParser struct {
	messageLength uint16
	index         uint16
	command       []byte
	commandType   CommandType
}

type ProtocolParserProducer struct {
}

func (t *ProtocolParserProducer) Produce() IProtocolParser {
	return &ProtocolParser{
		messageLength: 0,
		index:         0,
	}
}

// ParseStreamedData is to parse the byte data comes from server and convert it to meaningful internal commands
func (t *ProtocolParser) ParseStreamedData(readedByte byte) (interface{}, error) {

	// this is the first byte so we'll understand the command type here
	if t.index == index_CommandTypeStart {
		t.commandType = CommandType(readedByte)
	}

	t.command = append(t.command, readedByte)
	t.index = t.index + 1

	if t.index >= index_MessageLengthEnd {
		// 2nd and 3rd bytes are to store message length
		t.messageLength = binary.LittleEndian.Uint16(t.command[index_MessageLengthStart:index_MessageLengthEnd])
	} else {
		return nil, nil
	}

	// if we are not complete yet, just add byte to array and return
	if t.index < t.messageLength {
		return nil, nil
	}

	if t.commandType == CommandTypeWhoAmI {
		t.clearState()
		if t.messageLength == 0 {
			return WhoAmICommand{}, nil
		}

		return WhoAmICommand{
			ClientID: binary.LittleEndian.Uint64(t.command[index_MessageLengthEnd : index_MessageLengthEnd+8]),
		}, nil
	} else if t.commandType == CommandTypeListClients {
		var clientIDs []uint64

		for i := index_MessageLengthEnd + 8; i <= int(t.messageLength); i = i + 8 {
			clientIDs = append(clientIDs, binary.LittleEndian.Uint64(t.command[i-8:i]))
		}

		t.clearState()
		return ListClientsCommand{
			ConnectedClients: clientIDs,
		}, nil
	} else if t.commandType == CommandTypeMessageFromClient {
		sender := binary.LittleEndian.Uint64(t.command[index_MessageLengthEnd : index_MessageLengthEnd+8])
		body := t.command[index_MessageLengthEnd+8:]

		t.clearState()
		return MessageFromClient{
			SenderID: sender,
			Body:     body,
		}, nil
	} else if t.commandType == CommandTypeSendMessage {
		// if message type is send message we store the recipient count in 4th and 5th bytes
		recipientsCount := binary.LittleEndian.Uint16(t.command[index_RecipientsLengthStart:index_RecipientsLengthEnd])

		var recipients []uint64
		for i := index_RecipientsLengthEnd + 8; i < index_RecipientsLengthEnd+8+int(recipientsCount*8); i = i + 8 {
			recipients = append(recipients, binary.LittleEndian.Uint64(t.command[i-8:i]))
		}
		messageBody := t.command[index_RecipientsLengthEnd+(recipientsCount*8):]

		t.clearState()
		return SendMessageCommand{
			Recipients: recipients,
			Body:       messageBody,
		}, nil
	} else {
		t.clearState()
		return nil, errors.New("Unknown Command Processed")
	}

}

func (t *ProtocolParser) clearState() {
	// clear the array
	t.command = t.command[:0]
	t.index = 0
}
