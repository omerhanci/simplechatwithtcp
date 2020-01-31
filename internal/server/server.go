package server

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/Applifier/golang-backend-assignment/protocol"
)

type client struct {
	conn   net.Conn
	id     *uint64
	writer *bufio.Writer
}

type Server struct {
	listener       net.Listener
	writer         *bufio.Writer
	reader         *bufio.Reader
	clients        []*client
	mutex          *sync.Mutex
	sendMessage    chan protocol.SendMessageCommand
	whoami         chan protocol.WhoAmICommand
	listClients    chan protocol.ListClientsCommand
	protocolParser *protocol.ProtocolParser
}

func New() *Server {
	return &Server{
		sendMessage: make(chan protocol.SendMessageCommand),
		whoami:      make(chan protocol.WhoAmICommand),
		listClients: make(chan protocol.ListClientsCommand)}
}

func (server *Server) Start(laddr *net.TCPAddr) error {
	fmt.Println("TODO: Start handling client connections and messages")
	for {
		conn, err := server.listener.Accept()

		if err != nil {
			log.Print(err)
		} else {
			client := server.accept(conn)
			go server.serve(client)
		}

	}
}

func (server *Server) serve(client *client) {

	server.protocolParser = protocol.NewProtocolParser()
	// defer server.remove(client)

	for {
		data, err := server.reader.ReadByte()

		if err != nil {
			log.Printf("Read error %v", err)
		}

		command, err := server.protocolParser.ParseStreamedData(data)

		if err != nil {
			log.Printf("Parse error %v", err)
		}

		if command != nil {
			switch v := command.(type) {
			case protocol.WhoAmICommand:
				data := []byte{uint8(protocol.CommandTypeWhoAmI)}
				messageLengthBytes := make([]byte, 2)
				clientIDBytes := make([]byte, 8)

				binary.LittleEndian.PutUint64(messageLengthBytes, 11)
				binary.LittleEndian.PutUint64(clientIDBytes, *client.id)

				data = append(data, messageLengthBytes...)
				data = append(data, clientIDBytes...)
				client.writer.Write(data)
			case protocol.ListClientsCommand:
				data := []byte{uint8(protocol.CommandTypeListClients)}
				messageLengthBytes := make([]byte, 2)
				connectedClientBytes := make([]byte, 8)
				binary.LittleEndian.PutUint64(messageLengthBytes, uint64(len(server.clients)*8)+3)

				for i := 0; i < len(server.clients); i++ {
					binary.LittleEndian.PutUint64(connectedClientBytes, *server.clients[i].id)
					data = append(data, connectedClientBytes...)
				}
				client.writer.Write(data)

			case protocol.MessageFromClient:

			default:
				log.Printf("Unknown command: %v", v)
			}
		}
	}
}

func (server *Server) accept(conn net.Conn) *client {
	log.Printf("Accepting connection from %v, total clients: %v", conn.RemoteAddr().String(), len(server.clients)+1)

	server.mutex.Lock()
	defer server.mutex.Unlock()
	clientID := uint64(len(server.clients))

	client := &client{
		conn:   conn,
		id:     &clientID,
		writer: bufio.NewWriter(conn),
	}

	server.clients = append(server.clients, client)

	return client
}

func (server *Server) ListClientIDs() []uint64 {
	fmt.Println("TODO: Return the IDs of the connected clients")
	return []uint64{}
}

func (server *Server) Stop() error {
	fmt.Println("TODO: Stop accepting connections and close the existing ones")
	return nil
}
