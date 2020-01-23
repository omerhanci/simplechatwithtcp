package protocol

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
)

type CommandReader struct {
	reader *bufio.Reader
}

func NewCommandReader(reader io.Reader) *CommandReader {
	return &CommandReader{
		reader: bufio.NewReader(reader),
	}
}

func (r *CommandReader) Read() (interface{}, error) {
	// Read the first part
	commandName, err := r.reader.ReadString(' ')

	if err != nil {
		return nil, err
	}

	switch commandName {
	case "WhoAmI ":
		clientIDBytes, err := r.reader.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		var clientID uint64
		err = binary.Read(bytes.NewBuffer(clientIDBytes[:]), binary.LittleEndian, &clientID)

		if err != nil {
			return nil, err
		}
		return WhoAmICommand{
			ClientID: clientID,
		}, nil

	case "ListClients ":
		clientIDsBytes, err := r.reader.ReadBytes('\n')

		if err != nil {
			return nil, err
		}

		var clientIDs []uint64
		err = binary.Read(bytes.NewBuffer(clientIDsBytes[:]), binary.LittleEndian, &clientIDs)

		if err != nil {
			return nil, err
		}

		return ListClientsCommand{
			ConnectedClients: clientIDs,
		}, nil

	case "SendMessage ":
		recipientsBytes, err := r.reader.ReadBytes('\n')

		if err != nil {
			return nil, err
		}

		var recipients []uint64
		err = binary.Read(bytes.NewBuffer(recipientsBytes[:]), binary.LittleEndian, &recipients)

		if err != nil {
			return nil, err
		}

		messageBodyBytes, err := r.reader.ReadBytes('\n')

		if err != nil {
			return nil, err
		}

		var messageBody []byte
		err = binary.Read(bytes.NewBuffer(messageBodyBytes[:]), binary.LittleEndian, &messageBody)

		if err != nil {
			return nil, err
		}

		return SendMessageCommand{
			Recipients: recipients,
			Body:       messageBody,
		}, nil

	default:
		log.Printf("Unknown command: %v", commandName)
	}

	return nil, UnknownCommand
}

func (r *CommandReader) ReadAll() ([]interface{}, error) {
	commands := []interface{}{}

	for {
		command, err := r.Read()

		if command != nil {
			commands = append(commands, command)
		}

		if err == io.EOF {
			break
		} else if err != nil {
			return commands, err
		}
	}

	return commands, nil
}
