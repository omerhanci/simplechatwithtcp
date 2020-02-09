package channels

import (
	"errors"
	"log"

	"github.com/Applifier/golang-backend-assignment/protocol"
)

type CommandChannels struct {
	incoming    chan protocol.MessageFromClient
	whoami      chan protocol.WhoAmICommand
	listClients chan protocol.ListClientsCommand
}

type CommandChannelsProducer struct{}

func (t *CommandChannelsProducer) Produce() *CommandChannels {
	return &CommandChannels{
		incoming:    make(chan protocol.MessageFromClient),
		whoami:      make(chan protocol.WhoAmICommand),
		listClients: make(chan protocol.ListClientsCommand),
	}
}

func (t *CommandChannels) Add(data interface{}) error {
	switch v := data.(type) {
	case protocol.WhoAmICommand:
		t.whoami <- v
	case protocol.ListClientsCommand:
		t.listClients <- v
	case protocol.MessageFromClient:
		t.incoming <- v
	default:
		log.Printf("Unknown command: %v", v)
		return errors.New("Unknown command")
	}
	return nil
}

func (t *CommandChannels) Get(commandType protocol.CommandType) (interface{}, error) {
	switch commandType {
	case protocol.CommandTypeWhoAmI:
		return <-t.whoami, nil
	case protocol.CommandTypeListClients:
		return <-t.listClients, nil
	case protocol.CommandTypeMessageFromClient:
		return <-t.incoming, nil
	}
	return nil, errors.New("invalid command type")
}
