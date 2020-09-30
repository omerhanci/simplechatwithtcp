package protocol

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	fakeWhoAmICommand            = WhoAmICommand{ClientID: uint64(1)}
	fakeConnectedClientsCommand  = ListClientsCommand{ConnectedClients: []uint64{uint64(1)}}
	fakeMessageFromClientCommand = MessageFromClient{
		SenderID: uint64(1),
		Body:     []byte("hello"),
	}
	fakeSendMsgCommand = SendMessageCommand{
		Recipients: []uint64{uint64(1)},
		Body:       []byte("hello"),
	}
)

//1-0-0 should be whoami command
func TestWhoAmICommandShouldBeProduced(t *testing.T) {
	producer := ProtocolParserProducer{}
	protocolParser := producer.Produce()

	resp, err := protocolParser.ParseStreamedData(byte(1))
	assert.Nil(t, resp)
	assert.Nil(t, err)

	resp, err = protocolParser.ParseStreamedData(byte(0))
	assert.Nil(t, resp)
	assert.Nil(t, err)

	resp, err = protocolParser.ParseStreamedData(byte(0))
	fakeWhoAmICommand.ClientID = 0
	assert.Equal(t, resp, fakeWhoAmICommand)
	assert.Nil(t, err)
}

func TestWhoAmICommandShouldBeProducedWithCorrectID(t *testing.T) {
	producer := ProtocolParserProducer{}
	protocolParser := producer.Produce()

	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(11))

	clientIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(clientIDBytes, uint64(1))
	fakeWhoAmICommand.ClientID = 1

	resp, err := protocolParser.ParseStreamedData(byte(1))
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(messageLengthBytes[0])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(messageLengthBytes[1])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[0])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[1])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[2])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[3])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[4])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[5])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[6])
	assert.Nil(t, resp)
	assert.Nil(t, err)

	resp, err = protocolParser.ParseStreamedData(clientIDBytes[7])
	assert.Equal(t, resp, fakeWhoAmICommand)
	assert.Nil(t, err)

}

func TestListClientsCommandShouldBeProducedWithCorrectID(t *testing.T) {
	producer := ProtocolParserProducer{}
	protocolParser := producer.Produce()

	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(11))

	clientIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(clientIDBytes, uint64(1))

	resp, err := protocolParser.ParseStreamedData(byte(2))
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(messageLengthBytes[0])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(messageLengthBytes[1])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[0])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[1])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[2])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[3])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[4])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[5])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(clientIDBytes[6])
	assert.Nil(t, resp)
	assert.Nil(t, err)

	resp, err = protocolParser.ParseStreamedData(clientIDBytes[7])
	assert.Equal(t, resp, fakeConnectedClientsCommand)
	assert.Nil(t, err)

}

func TestMessageFromClientsCommandShouldBeProduced(t *testing.T) {
	producer := ProtocolParserProducer{}
	protocolParser := producer.Produce()

	senderIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(senderIDBytes, uint64(1))

	body := []byte("hello")

	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(len(body)+11))

	resp, err := protocolParser.ParseStreamedData(byte(4))
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(messageLengthBytes[0])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(messageLengthBytes[1])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(senderIDBytes[0])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(senderIDBytes[1])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(senderIDBytes[2])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(senderIDBytes[3])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(senderIDBytes[4])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(senderIDBytes[5])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(senderIDBytes[6])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(senderIDBytes[7])
	assert.Nil(t, resp)
	assert.Nil(t, err)

	for i := 0; i < len(body)-1; i++ {
		resp, err = protocolParser.ParseStreamedData(body[i])
		assert.Nil(t, resp)
		assert.Nil(t, err)
	}

	resp, err = protocolParser.ParseStreamedData(body[len(body)-1])
	assert.Equal(t, fakeMessageFromClientCommand, resp)
	assert.Nil(t, err)

}

func TestSendMessageCommandShouldBeProduced(t *testing.T) {
	producer := ProtocolParserProducer{}
	protocolParser := producer.Produce()

	recipientIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(recipientIDBytes, uint64(1))

	body := []byte("hello")

	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(len(body)+13))

	recipientsLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(recipientsLengthBytes, uint16(1))

	resp, err := protocolParser.ParseStreamedData(byte(3))
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(messageLengthBytes[0])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(messageLengthBytes[1])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(recipientsLengthBytes[0])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(recipientsLengthBytes[1])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(recipientIDBytes[0])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(recipientIDBytes[1])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(recipientIDBytes[2])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(recipientIDBytes[3])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(recipientIDBytes[4])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(recipientIDBytes[5])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(recipientIDBytes[6])
	assert.Nil(t, resp)
	assert.Nil(t, err)
	resp, err = protocolParser.ParseStreamedData(recipientIDBytes[7])
	assert.Nil(t, resp)
	assert.Nil(t, err)

	for i := 0; i < len(body)-1; i++ {
		resp, err = protocolParser.ParseStreamedData(body[i])
		assert.Nil(t, resp)
		assert.Nil(t, err)
	}

	resp, err = protocolParser.ParseStreamedData(body[len(body)-1])
	assert.Equal(t, fakeSendMsgCommand, resp)
	assert.Nil(t, err)

}

func TestUnknownCommandShouldReturnNil(t *testing.T) {
	producer := ProtocolParserProducer{}
	protocolParser := producer.Produce()

	resp, err := protocolParser.ParseStreamedData(byte(5))
	assert.Nil(t, resp)
	assert.Nil(t, err)

	resp, err = protocolParser.ParseStreamedData(byte(0))
	assert.Nil(t, resp)
	assert.Nil(t, err)

	resp, err = protocolParser.ParseStreamedData(byte(0))
	assert.Nil(t, resp)
	assert.Error(t, err)

}

func TestWhoAmICommandToByteArray(t *testing.T) {
	commandBytes := []byte{}

	command := WhoAmICommand{ClientID: 1}
	convertedBytes := command.ToByteArray()

	commandTypeBytes := []byte{uint8(CommandTypeWhoAmI)}

	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(11))

	clientIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(clientIDBytes, uint64(1))

	commandBytes = append(commandBytes, commandTypeBytes...)
	commandBytes = append(commandBytes, messageLengthBytes...)
	commandBytes = append(commandBytes, clientIDBytes...)

	assert.Equal(t, commandBytes, convertedBytes)
}

func TestListClientsToByteArray(t *testing.T) {
	commandBytes := []byte{}

	command := ListClientsCommand{ConnectedClients: []uint64{uint64(1), uint64(2)}}
	convertedBytes := command.ToByteArray(uint64(2))

	commandTypeBytes := []byte{uint8(CommandTypeListClients)}

	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(11))

	clientIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(clientIDBytes, uint64(1))

	commandBytes = append(commandBytes, commandTypeBytes...)
	commandBytes = append(commandBytes, messageLengthBytes...)
	commandBytes = append(commandBytes, clientIDBytes...)

	assert.Equal(t, commandBytes, convertedBytes)
}

func TestMessageFromClientToByteArray(t *testing.T) {
	commandBytes := []byte{}

	command := fakeMessageFromClientCommand
	convertedBytes := command.ToByteArray()

	commandTypeBytes := []byte{uint8(CommandTypeMessageFromClient)}

	senderIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(senderIDBytes, uint64(1))

	body := []byte("hello")

	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(len(body)+11))

	commandBytes = append(commandBytes, commandTypeBytes...)
	commandBytes = append(commandBytes, messageLengthBytes...)
	commandBytes = append(commandBytes, senderIDBytes...)
	commandBytes = append(commandBytes, body...)

	assert.Equal(t, commandBytes, convertedBytes)
}

func TestSendMessageToByteArray(t *testing.T) {
	commandBytes := []byte{}
	command := fakeSendMsgCommand
	convertedBytes := command.ToByteArray()

	commandTypeBytes := []byte{uint8(CommandTypeSendMessage)}

	recipientIDBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(recipientIDBytes, uint64(1))

	body := []byte("hello")

	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(len(body)+13))

	recipientsLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(recipientsLengthBytes, uint16(1))

	commandBytes = append(commandBytes, commandTypeBytes...)
	commandBytes = append(commandBytes, messageLengthBytes...)
	commandBytes = append(commandBytes, recipientsLengthBytes...)
	commandBytes = append(commandBytes, recipientIDBytes...)
	commandBytes = append(commandBytes, body...)

	assert.Equal(t, commandBytes, convertedBytes)
}

func TestCreateQueryCommand(t *testing.T) {
	commandBytes := []byte{}

	command := QueryCommand{}
	convertedBytes := command.CreateQueryCommand(CommandTypeWhoAmI)

	commandTypeBytes := []byte{uint8(CommandTypeWhoAmI)}

	messageLengthBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(messageLengthBytes, uint16(0))

	commandBytes = append(commandBytes, commandTypeBytes...)
	commandBytes = append(commandBytes, messageLengthBytes...)

	assert.Equal(t, commandBytes, convertedBytes)
}
