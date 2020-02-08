package server

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"sync"

	"github.com/Applifier/golang-backend-assignment/datastream"

	"github.com/Applifier/golang-backend-assignment/protocol"
)

// client struct is to hold connected client data internally
type client struct {
	dataStreamer datastream.IDataStreamer
	id           *uint64
}

// Server struct
type Server struct {
	dataStreamer   datastream.IDataStreamer
	clients        []*client
	clientIDs      []uint64
	clientMutex    *sync.Mutex
	sendMessage    chan protocol.SendMessageCommand
	whoami         chan protocol.WhoAmICommand
	listClients    chan protocol.ListClientsCommand
	protocolParser *protocol.ProtocolParser
}

// New is to create new server and return
func New() *Server {
	return &Server{
		sendMessage: make(chan protocol.SendMessageCommand),
		whoami:      make(chan protocol.WhoAmICommand),
		listClients: make(chan protocol.ListClientsCommand),
		clientMutex: &sync.Mutex{}}
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

// Start function starts the server and make it ready to accept connections
func (server *Server) Start(laddr *net.TCPAddr) error {
	dataStreamerProducer := datastream.TcpDataStreamProducer{}
	dataStreamer := dataStreamerProducer.Produce()
	listener, err := dataStreamer.CreateListener(laddr)

	if err != nil {
		log.Print(err)
		return err
	}
	server.dataStreamer = listener

	go server.listen()
	return nil

}

// serve function is to read streamed data from connected client and turn it to meaningful commands
func (server *Server) serve(client *client) {

	server.protocolParser = protocol.NewProtocolParser()
	defer server.remove(client)

	for {
		data, err := client.dataStreamer.ReadByte()

		if err == io.EOF {
			server.Stop()
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
				data := []byte{uint8(protocol.CommandTypeWhoAmI)}
				messageLengthBytes := make([]byte, 2)
				clientIDBytes := make([]byte, 8)

				binary.LittleEndian.PutUint16(messageLengthBytes, 11)
				binary.LittleEndian.PutUint64(clientIDBytes, *client.id)

				data = append(data, messageLengthBytes...)
				data = append(data, clientIDBytes...)
				client.dataStreamer.Write(data)
				client.dataStreamer.Flush()
			case protocol.ListClientsCommand:
				data := []byte{uint8(protocol.CommandTypeListClients)}
				messageLengthBytes := make([]byte, 2)
				connectedClientBytes := make([]byte, 8)
				binary.LittleEndian.PutUint16(messageLengthBytes, uint16(((len(server.clients)-1)*8)+3))
				data = append(data, messageLengthBytes...)
				for i := 0; i < len(server.clients); i++ {
					if *server.clients[i].id != *client.id {
						binary.LittleEndian.PutUint64(connectedClientBytes, *server.clients[i].id)
						data = append(data, connectedClientBytes...)
					}
				}
				client.dataStreamer.Write(data)
				client.dataStreamer.Flush()

			case protocol.SendMessageCommand:
				data := []byte{uint8(protocol.CommandTypeMessageFromClient)}
				messageLengthBytes := make([]byte, 2)
				senderBytes := make([]byte, 8)
				sendMessageCommand := command.(protocol.SendMessageCommand)
				bodyBytes := sendMessageCommand.Body
				binary.LittleEndian.PutUint64(senderBytes, *client.id)
				binary.LittleEndian.PutUint16(messageLengthBytes, uint16(len(bodyBytes)+11))
				data = append(data, messageLengthBytes...)
				data = append(data, senderBytes...)
				data = append(data, bodyBytes...)
				for i := 0; i < len(sendMessageCommand.Recipients); i++ {
					client := server.getClientByID(sendMessageCommand.Recipients[i])
					client.dataStreamer.Write(data)
					client.dataStreamer.Flush()
				}

			default:
				log.Printf("Unknown command: %v", v)
			}
		}
	}
}

// getClientById gets client id and returns id with that client
func (server *Server) getClientByID(clientID uint64) *client {
	for i := 0; i < len(server.clients); i++ {
		if *server.clients[i].id == clientID {
			return server.clients[i]
		}
	}
	return nil
}

// accept connections from connected client add it to connected clients
func (server *Server) accept(clientStreamer datastream.IDataStreamer) *client {
	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
	clientID := uint64(len(server.clients) + 1)

	client := &client{
		dataStreamer: clientStreamer,
		id:           &clientID,
	}

	server.clients = append(server.clients, client)
	server.clientIDs = append(server.clientIDs, *client.id)

	return client
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
