package client

import (
	"encoding/binary"
	"io"
	"log"
	"net"

	"github.com/Applifier/golang-backend-assignment/datastream"

	"github.com/Applifier/golang-backend-assignment/protocol"
)

// Client structure
type Client struct {
	dataStream     datastream.IDataStreamer
	incoming       chan protocol.MessageFromClient
	whoami         chan protocol.WhoAmICommand
	listClients    chan protocol.ListClientsCommand
	protocolParser *protocol.ProtocolParser
}

// New is to create new client and return
func New() *Client {
	return &Client{
		incoming:       make(chan protocol.MessageFromClient),
		whoami:         make(chan protocol.WhoAmICommand),
		listClients:    make(chan protocol.ListClientsCommand),
		protocolParser: protocol.NewProtocolParser(),
	}
}

// Connect function is to connect to server given serverAddr parameter
func (cli *Client) Connect(serverAddr *net.TCPAddr) error {

	dataStreamerProducer := datastream.TcpDataStreamProducer{}
	dataStreamer := dataStreamerProducer.Produce()
	tcpDataStreamer, err := dataStreamer.CreateConnection(serverAddr)

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
			cli.Close()
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
			switch v := command.(type) {
			case protocol.WhoAmICommand:
				cli.whoami <- v
			case protocol.ListClientsCommand:
				cli.listClients <- v
			case protocol.MessageFromClient:
				cli.incoming <- v
			default:
				log.Printf("Unknown command: %v", v)
			}
		}
	}
}

// Close the connection
func (cli *Client) Close() error {
	err := cli.dataStream.CloseConnection()
	if err != nil {
		log.Printf("cannot close: %v", err)
	}
	return nil
}

// WhoAmI function is to get the client id from the server
func (cli *Client) WhoAmI() (uint64, error) {
	// send a whoami message to the server then wait for response to come to the channel
	cli.sendMessageToServer(uint8(protocol.CommandTypeWhoAmI), 0, nil)
	cmdResponse := <-cli.whoami
	return cmdResponse.ClientID, nil
}

// ListClientIDs function is to get current connected clients' ids from the server
func (cli *Client) ListClientIDs() ([]uint64, error) {
	// send a listClients message to the server then wait for response to come to the channel
	cli.sendMessageToServer(uint8(protocol.CommandTypeListClients), 0, nil)
	cmdResponse := <-cli.listClients
	return cmdResponse.ConnectedClients, nil
}

// SendMsg function is to Send messages to the other connected clients
func (cli *Client) SendMsg(recipients []uint64, body []byte) error {
	dataBytes := []byte{}
	// we have a special case here, 4th and 5th bytes for recipient length, send them as data
	recipientLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(recipientLengthBytes, uint16(len(recipients)))
	dataBytes = append(dataBytes, recipientLengthBytes...)

	// add recipients to the data
	recipientIDBytes := make([]byte, 8)
	for i := 0; i < len(recipients); i++ {
		binary.LittleEndian.PutUint64(recipientIDBytes, recipients[i])
		dataBytes = append(dataBytes, recipientIDBytes...)
	}

	// and finally add the message body
	dataBytes = append(dataBytes, body...)

	// calculate message length
	// commandType + messageLength + recipientsLength + recipients + messageBody
	messageLength := protocol.CommandLengthType + protocol.CommandLengthMessageLength + protocol.CommandLengthRecipientsLength + (len(recipients) * 8) + len(body)

	cli.sendMessageToServer(uint8(protocol.CommandTypeSendMessage), uint16(messageLength), dataBytes)
	return nil
}

// HandleIncomingMessages function is to get messages from the other clients and push it to the channels
func (cli *Client) HandleIncomingMessages(writeCh chan<- protocol.MessageFromClient) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("run time panic: %v", err)
			cli.Close()
		}
	}()
	for {
		message := <-cli.incoming
		writeCh <- message
	}

}

func (cli *Client) sendMessageToServer(commandType uint8, messageLength uint16, data []byte) {
	// first byte is for command type
	command := []byte{commandType}
	// 2nd and 3rd bytes for messageLength
	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, messageLength)
	command = append(command, messageLengthBytes...)
	// rest is data
	if data != nil {
		command = append(command, data...)
	}

	//todo error handling
	cli.dataStream.Write(command)
	cli.dataStream.Flush()
}
