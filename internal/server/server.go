package server

import (
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

func (server *Server) listen() error {
	defer server.Stop()
	for {
		clientStreamer, err := server.dataStreamer.Accept()

		if err != nil {
			log.Print(err)
			return err
		} else {
			client := server.createClient(clientStreamer)
			go server.serve(client)
		}

	}
}

func (server *Server) createClient(clientStreamer datastream.IDataStreamer) *client {
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

	defer server.remove(client)

	for {
		data, err := client.dataStreamer.ReadByte()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Printf("Read error %v", err)
			break
		}

		command, err := protocolParser.ParseStreamedData(data)

		if err != nil {
			log.Printf("Parse error %v", err)
			break
		}

		if command != nil {
			switch v := command.(type) {
			case protocol.WhoAmICommand:
				server.handleWhoAmICommand(client)
				break
			case protocol.ListClientsCommand:
				server.handleListClientsCommand(client)
				break
			case protocol.SendMessageCommand:
				server.handleSendMessageCommand(client, command.(protocol.SendMessageCommand))
				break
			default:
				log.Printf("Unknown command: %v", v)
				break
			}
		}
	}
}

// Stop accepting connections and close the existing ones
func (server *Server) Stop() error {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
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

// ListClientIDs return the connected clients ids
func (server *Server) ListClientIDs() []uint64 {
	return server.clientIDs
}

func (server *Server) sendMessageToClient(client *client, message []byte) {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
	client.dataStreamer.Write(message)
	client.dataStreamer.Flush()
}

func (server *Server) handleWhoAmICommand(client *client) {
	command := protocol.WhoAmICommand{ClientID: client.id}
	server.sendMessageToClient(client, command.ToByteArray())
}

func (server *Server) handleListClientsCommand(client *client) {
	command := protocol.ListClientsCommand{ConnectedClients: server.clientIDs}
	server.sendMessageToClient(client, command.ToByteArray(client.id))
}

func (server *Server) handleSendMessageCommand(client *client, command protocol.SendMessageCommand) {
	msgFromClientCommand := protocol.MessageFromClient{Body: command.Body, SenderID: client.id}
	for i := 0; i < len(command.Recipients); i++ {
		recipient := server.getClientByID(command.Recipients[i])
		server.sendMessageToClient(recipient, msgFromClientCommand.ToByteArray())
	}
}

func (server *Server) getClientByID(clientID uint64) *client {
	for i := 0; i < len(server.clients); i++ {
		if server.clients[i].id == clientID {
			return server.clients[i]
		}
	}
	return nil
}
