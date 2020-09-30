package client

import (
	"io"
	"log"
	"net"

	"github.com/Applifier/golang-backend-assignment/channels"

	"github.com/Applifier/golang-backend-assignment/datastream"

	"github.com/Applifier/golang-backend-assignment/protocol"
)

// Client structure
type Client struct {
	dataStream      datastream.IDataStreamer
	commandChannels channels.ICommandChannels
	protocolParser  protocol.IProtocolParser
}

// New is to create new client and return
func New() *Client {
	dataStreamerProducer := datastream.TcpDataStreamProducer{}
	dataStreamer := dataStreamerProducer.Produce()

	protocolParserProducer := protocol.ProtocolParserProducer{}
	protocolParser := protocolParserProducer.Produce()

	commandChannelsProducer := channels.CommandChannelsProducer{}
	commandChannels := commandChannelsProducer.Produce()
	return &Client{
		dataStream:      dataStreamer,
		commandChannels: commandChannels,
		protocolParser:  protocolParser,
	}
}

// Connect function is to connect to server given serverAddr parameter
func (cli *Client) Connect(serverAddr *net.TCPAddr) error {
	tcpDataStreamer, err := cli.dataStream.CreateConnection(serverAddr)
	cli.dataStream = tcpDataStreamer

	if err != nil {
		return err
	}

	go cli.Start()
	return nil

}

// Start function is to Start Reading from tcp connection and send the data to related channels
func (cli *Client) Start() {
	for {
		data, err := cli.dataStream.ReadByte()

		// server is closed, we should close the connection
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Read error %v", err)
			break
		}

		// parse the streamed data to make it meaningful
		command, err := cli.protocolParser.ParseStreamedData(data)

		if err != nil {
			log.Printf("Parse error %v", err)
			break
		}

		// if command is not nil, then we have a comlete command object, send it to the related channels
		if command != nil {
			cli.commandChannels.Add(command)
		}
	}
}

// Close the connection
func (cli *Client) Close() error {
	err := cli.dataStream.CloseConnection()
	if err != nil {
		log.Printf("cannot close: %v", err)
		return err
	}
	return nil
}

// WhoAmI function is to get the client id from the server
func (cli *Client) WhoAmI() (uint64, error) {
	command := protocol.QueryCommand{}
	// send a whoami message to the server then wait for response to come to the channel
	err := cli.sendMessageToServer(command.CreateQueryCommand(protocol.CommandTypeWhoAmI))
	if err != nil {
		return 0, err
	}
	cmdResponse, err := cli.commandChannels.Get(protocol.CommandTypeWhoAmI)
	if err != nil {
		return 0, err
	}
	clientID := cmdResponse.(protocol.WhoAmICommand).ClientID
	return clientID, nil
}

// ListClientIDs function is to get current connected clients' ids from the server
func (cli *Client) ListClientIDs() ([]uint64, error) {
	command := protocol.QueryCommand{}
	// send a listClients message to the server then wait for response to come to the channel
	err := cli.sendMessageToServer(command.CreateQueryCommand(protocol.CommandTypeListClients))
	if err != nil {
		return nil, err
	}
	cmdResponse, err := cli.commandChannels.Get(protocol.CommandTypeListClients)
	if err != nil {
		return nil, err
	}
	connectedClients := cmdResponse.(protocol.ListClientsCommand).ConnectedClients
	return connectedClients, nil
}

// SendMsg function is to Send messages to the other connected clients
func (cli *Client) SendMsg(recipients []uint64, body []byte) error {
	command := protocol.SendMessageCommand{Recipients: recipients, Body: body}
	err := cli.sendMessageToServer(command.ToByteArray())
	if err != nil {
		return err
	}
	return nil
}

// HandleIncomingMessages function is to get messages from the other clients and push it to the channels
func (cli *Client) HandleIncomingMessages(writeCh chan<- protocol.MessageFromClient) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("run time panic: %v", err)
		}
	}()
	for {
		message, err := cli.commandChannels.Get(protocol.CommandTypeMessageFromClient)
		if err != nil {
			log.Printf("Cannot get message from client: %v", err)
			break
		}
		writeCh <- message.(protocol.MessageFromClient)
	}

}

func (cli *Client) sendMessageToServer(data []byte) error {
	_, err := cli.dataStream.Write(data)
	if err != nil {
		return err
	}

	err = cli.dataStream.Flush()
	if err != nil {
		return err
	}

	return nil
}
