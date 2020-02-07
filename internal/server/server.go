package server

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/Applifier/golang-backend-assignment/protocol"
)

type client struct {
	conn   net.Conn
	id     *uint64
	writer *bufio.Writer
	reader *bufio.Reader
}

type Server struct {
	listener       net.Listener
	writer         *bufio.Writer
	reader         *bufio.Reader
	clients        []*client
	clientIDs      []uint64
	clientMutex    *sync.Mutex
	sendMessage    chan protocol.SendMessageCommand
	whoami         chan protocol.WhoAmICommand
	listClients    chan protocol.ListClientsCommand
	protocolParser *protocol.ProtocolParser
}

func New() *Server {
	return &Server{
		sendMessage: make(chan protocol.SendMessageCommand),
		whoami:      make(chan protocol.WhoAmICommand),
		listClients: make(chan protocol.ListClientsCommand),
		clientMutex: &sync.Mutex{}}
}

func (server *Server) listen(listener net.Listener) error {
	server.listener = listener
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

func (server *Server) Start(laddr *net.TCPAddr) error {
	fmt.Println("Start handling client connections and messages")

	listener, err := net.Listen("tcp", laddr.String())
	if err != nil {
		log.Print(err)
		return err
	}

	go server.listen(listener)
	return nil

}

func (server *Server) serve(client *client) {

	server.protocolParser = protocol.NewProtocolParser()
	defer server.remove(client)

	for {
		data, err := client.reader.ReadByte()

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
				client.writer.Write(data)
				client.writer.Flush()
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
				client.writer.Write(data)
				client.writer.Flush()

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
					client := server.getClientById(sendMessageCommand.Recipients[i])
					client.writer.Write(data)
					client.writer.Flush()
				}

			default:
				log.Printf("Unknown command: %v", v)
			}
		}
	}
}

func (server *Server) getClientById(clientID uint64) *client {
	for i := 0; i < len(server.clients); i++ {
		if *server.clients[i].id == clientID {
			return server.clients[i]
		}
	}
	return nil
}

func (server *Server) accept(conn net.Conn) *client {
	// log.Printf("Accepting connection from %v, total clients: %v", conn.RemoteAddr().String(), len(server.clients)+1)

	server.clientMutex.Lock()
	defer server.clientMutex.Unlock()
	clientID := uint64(len(server.clients) + 1)

	client := &client{
		conn:   conn,
		id:     &clientID,
		writer: bufio.NewWriter(conn),
		reader: bufio.NewReader(conn),
	}

	server.clients = append(server.clients, client)
	server.clientIDs = append(server.clientIDs, *client.id)

	return client
}

func (server *Server) ListClientIDs() []uint64 {
	// fmt.Println("Return the IDs of the connected clients")
	return server.clientIDs
}

func (server *Server) Stop() error {
	// fmt.Println("TODO: Stop accepting connections and close the existing ones")
	server.listener.Close()
	for i := 0; i < len(server.clients); i++ {
		server.clients[i].conn.Close()
	}
	return nil
}

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

	// log.Printf("Closing connection from %v", client.conn.RemoteAddr())
	client.conn.Close()
}
