package server

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"

	"github.com/Applifier/golang-backend-assignment/channels"

	"github.com/Applifier/golang-backend-assignment/datastream"

	"github.com/Applifier/golang-backend-assignment/protocol"
)

// client struct is to hold connected client data internally
type client struct {
	dataStreamer datastream.IDataStreamer
	id           uint64
}

// Server struct
type Server struct {
	dataStreamer           datastream.IDataStreamer
	clients                []*client
	clientIDs              []uint64
	clientMutex            *sync.Mutex
	commandChannels        channels.ICommandChannels
	protocolParser         protocol.IProtocolParser
	protocolParserProducer protocol.IProtocolParserProducer
}

// New is to create new server and return
func New() *Server {
	commandChannelsProducer := channels.CommandChannelsProducer{}
	commandChannels := commandChannelsProducer.Produce()

	dataStreamerProducer := datastream.TcpDataStreamProducer{}
	dataStreamer := dataStreamerProducer.Produce()

	protocolParserProducer := &protocol.ProtocolParserProducer{}

	return &Server{
		commandChannels:        commandChannels,
		clientMutex:            &sync.Mutex{},
		protocolParserProducer: protocolParserProducer,
		dataStreamer:           dataStreamer}
}

// Start function starts the server and make it ready to accept connections
func (server *Server) Start(laddr *net.TCPAddr) error {

	listener, err := server.dataStreamer.CreateListener(laddr)
	if err != nil {
		log.Print(err)
		return err
	}
	server.dataStreamer = listener

	go server.listen()
	return nil

}

// listen function listens and accepts client connections
func (server *Server) listen() error {
	for {
		clientStreamer, err := server.dataStreamer.Accept()

		if err != nil {
			log.Print(err)
		} else {
			client := server.accept(clientStreamer)
			go server.serve(client)
		}

	}
}

// accept connections from connected client add it to connected clients
func (server *Server) accept(clientStreamer datastream.IDataStreamer) *client {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
	clientID := uint64(len(server.clients) + 1)

	client := &client{
		dataStreamer: clientStreamer,
		id:           clientID,
	}

	server.clients = append(server.clients, client)
	server.clientIDs = append(server.clientIDs, client.id)

	return client
}

// serve function is to read streamed data from connected client and turn it to meaningful commands
func (server *Server) serve(client *client) {

	// create new parser for this client only
	protocolParser := server.protocolParserProducer.Produce()
	server.protocolParser = protocolParser

	defer server.remove(client)

	for {
		data, err := client.dataStreamer.ReadByte()

		if err == io.EOF {
			log.Printf("EOF reached")
			break
		}

		if err != nil {
			log.Printf("Read error %v", err)
			break
		}

		command, err := server.protocolParser.ParseStreamedData(data)

		if err != nil {
			log.Printf("Parse error %v", err)
			break
		}

		if command != nil {
			switch v := command.(type) {
			case protocol.WhoAmICommand:
				server.handleWhoAmICommand(client)
			case protocol.ListClientsCommand:
				server.handleListClientsCommand(client)
			case protocol.SendMessageCommand:
				server.handleSendMessageCommand(client, command.(protocol.SendMessageCommand))

			default:
				log.Printf("Unknown command: %v", v)
			}
		}
	}
}

// ListClientIDs return the connected clients ids
func (server *Server) ListClientIDs() []uint64 {
	return server.clientIDs
}

// Stop accepting connections and close the existing ones
func (server *Server) Stop() error {
	server.dataStreamer.CloseListener()
	for i := 0; i < len(server.clients); i++ {
		server.clients[i].dataStreamer.CloseConnection()
	}
	return nil
}

// remove the connected client
func (server *Server) remove(client *client) {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()

	// remove the connections from the clients array
	for i, check := range server.clients {
		if check == client {
			server.clients = append(server.clients[:i], server.clients[i+1:]...)
			server.clientIDs = append(server.clientIDs[:i], server.clientIDs[i+1:]...)
		}
	}

	client.dataStreamer.CloseConnection()
}

func (server *Server) sendMessageToClient(client *client, commandType protocol.CommandType, messageLength int, data []byte) {
	// first byte is for command type
	command := []byte{uint8(commandType)}
	// 2nd and 3rd bytes for messageLength
	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(messageLength))
	command = append(command, messageLengthBytes...)
	// rest is data
	if data != nil {
		command = append(command, data...)
	}

	//todo error handling
	client.dataStreamer.Write(command)
	client.dataStreamer.Flush()
}

func (server *Server) handleWhoAmICommand(client *client) {
	dataBytes := []byte{}

	clientIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(clientIDBytes, client.id)
	dataBytes = append(dataBytes, clientIDBytes...)

	// calculate command length (commandType + messageLength + clientID)
	messageLength := protocol.CommandLengthType + protocol.CommandLengthMessageLength + 8
	server.sendMessageToClient(client, protocol.CommandTypeWhoAmI, messageLength, dataBytes)
}

func (server *Server) handleListClientsCommand(client *client) {
	dataBytes := []byte{}

	connectedClientBytes := make([]byte, 8)
	for i := 0; i < len(server.clients); i++ {
		if server.clients[i].id != client.id {
			binary.LittleEndian.PutUint64(connectedClientBytes, server.clients[i].id)
			dataBytes = append(dataBytes, connectedClientBytes...)
		}
	}

	// calculate command length (commandType + messageLength + (clientCount-1)*8)
	messageLength := protocol.CommandLengthType + protocol.CommandLengthMessageLength + (8 * (len(server.clients) - 1))
	server.sendMessageToClient(client, protocol.CommandTypeListClients, messageLength, dataBytes)
}

func (server *Server) handleSendMessageCommand(client *client, command protocol.SendMessageCommand) {
	dataBytes := []byte{}

	senderBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(senderBytes, client.id)

	bodyBytes := command.Body

	dataBytes = append(dataBytes, senderBytes...)
	dataBytes = append(dataBytes, bodyBytes...)

	// calculate command length (commandType + messageLength + senderid + messagebody)
	messageLength := protocol.CommandLengthType + protocol.CommandLengthMessageLength + protocol.CommandLengthClient + len(bodyBytes)
	for i := 0; i < len(command.Recipients); i++ {
		client := server.getClientByID(command.Recipients[i])
		server.sendMessageToClient(client, protocol.CommandTypeMessageFromClient, messageLength, dataBytes)
	}
}

// getClientById gets client id and returns id with that client
func (server *Server) getClientByID(clientID uint64) *client {
	for i := 0; i < len(server.clients); i++ {
		if server.clients[i].id == clientID {
			return server.clients[i]
		}
	}
	return nil
}
