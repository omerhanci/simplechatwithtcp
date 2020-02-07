package client

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/Applifier/golang-backend-assignment/protocol"
)

type IncomingMessage struct {
	SenderID uint64
	Body     []byte
}

type Client struct {
	conn           net.Conn
	writer         *bufio.Writer
	reader         *bufio.Reader
	incoming       chan protocol.MessageFromClient
	whoami         chan protocol.WhoAmICommand
	listClients    chan protocol.ListClientsCommand
	protocolParser *protocol.ProtocolParser
}

func New() *Client {
	return &Client{
		incoming:       make(chan protocol.MessageFromClient),
		whoami:         make(chan protocol.WhoAmICommand),
		listClients:    make(chan protocol.ListClientsCommand),
		protocolParser: protocol.NewProtocolParser(),
	}
}

func (cli *Client) Connect(serverAddr *net.TCPAddr) error {
	// fmt.Println("TODO: Connect to the server using the given address")
	conn, err := net.Dial("tcp", serverAddr.String())

	if err != nil {
		return err
	}

	cli.conn = conn
	cli.reader = bufio.NewReader(conn)
	cli.writer = bufio.NewWriter(conn)
	go cli.Start()
	return nil

}

func (cli *Client) Start() {
	for {
		data, err := cli.reader.ReadByte()

		if err == io.EOF {
			cli.Close()
			break
		}

		if err != nil {
			log.Printf("Read error %v", err)
			break
		}

		command, err := cli.protocolParser.ParseStreamedData(data)

		if err != nil {
			log.Printf("Parse error %v", err)
			break
		}

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

func (cli *Client) Close() error {
	// fmt.Println("Close the connection to the server")
	err := cli.conn.Close()
	if err != nil {
		log.Printf("cannot close: %v", err)
	}
	return nil
}

func (cli *Client) WhoAmI() (uint64, error) {
	// fmt.Println("Fetch the ID from the server")
	messageLength := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLength, 0)
	data := []byte{uint8(protocol.CommandTypeWhoAmI)}
	data = append(data, messageLength...)
	cli.writer.Write(data)
	cli.writer.Flush()
	cmdResponse := <-cli.whoami
	fmt.Println(cmdResponse.ClientID)
	return cmdResponse.ClientID, nil
}

func (cli *Client) ListClientIDs() ([]uint64, error) {
	// fmt.Println("TODO: Fetch the IDs from the server")
	messageLength := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLength, 0)
	data := []byte{uint8(protocol.CommandTypeListClients)}
	data = append(data, messageLength...)
	cli.writer.Write(data)
	cli.writer.Flush()
	cmdResponse := <-cli.listClients
	fmt.Println(cmdResponse.ConnectedClients)
	return cmdResponse.ConnectedClients, nil
}

func (cli *Client) SendMsg(recipients []uint64, body []byte) error {
	// fmt.Println("TODO: Send the message to the server")
	data := []byte{uint8(protocol.CommandTypeSendMessage)}
	messageLength := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLength, uint16(5+(len(recipients)*8)+len(body)))
	data = append(data, messageLength...)
	recipientLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(recipientLengthBytes, uint16(len(recipients)))
	data = append(data, recipientLengthBytes...)
	recipientIDBytes := make([]byte, 8)

	for i := 0; i < len(recipients); i++ {
		binary.LittleEndian.PutUint64(recipientIDBytes, recipients[i])
		data = append(data, recipientIDBytes...)
	}

	data = append(data, body...)
	cli.writer.Write(data)
	cli.writer.Flush()
	return nil
}

func (cli *Client) HandleIncomingMessages(writeCh chan<- protocol.MessageFromClient) {
	// fmt.Println("TODO: Handle the messages from the server")
	for {
		message := <-cli.incoming
		writeCh <- message
	}

}
