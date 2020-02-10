package client

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/Applifier/golang-backend-assignment/channels"
	"github.com/Applifier/golang-backend-assignment/datastream"
	"github.com/Applifier/golang-backend-assignment/protocol"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	fakeAddress                  = net.TCPAddr{Port: 1234}
	fakeWhoAmICommand            = protocol.WhoAmICommand{ClientID: uint64(1)}
	fakeConnectedClientsCommand  = protocol.ListClientsCommand{ConnectedClients: []uint64{uint64(1)}}
	fakeMessageFromClientCommand = protocol.MessageFromClient{
		SenderID: uint64(1),
		Body:     []byte("hello"),
	}
	fakeSendMsgCommand = protocol.SendMessageCommand{
		Recipients: []uint64{uint64(1)},
		Body:       []byte("hello"),
	}
)

func TestNewMethodShouldCreateNewClientInstance(t *testing.T) {
	client := New()
	assert.NotNil(t, client)
	assert.NotNil(t, client.commandChannels)
	assert.NotNil(t, client.protocolParser)
}

func TestConnectShouldReturnNoError(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)

	fakeDataStreamer.On("CreateConnection", mock.Anything).Return(fakeDataStreamer, nil).Once()
	fakeDataStreamer.On("ReadByte").Return(byte(0), nil).Maybe()
	client := New()
	client.dataStream = fakeDataStreamer
	response := client.Connect(&fakeAddress)
	assert.Nil(t, response)
}

func TestConnectShouldReturnErrorIfErrorOccured(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)

	fakeDataStreamer.On("CreateConnection", mock.Anything).Return(fakeDataStreamer, errors.New("cannot connect")).Once()
	fakeDataStreamer.On("ReadByte").Return(byte(0), nil).Maybe()
	client := New()
	client.dataStream = fakeDataStreamer

	response := client.Connect(&fakeAddress)
	assert.EqualError(t, response, "cannot connect")
}

func TestStartFunctionShouldStopAndCloseTheConnectionIfEOFReturns(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)

	fakeDataStreamer.On("ReadByte").Return(byte(0), io.EOF).Once()
	fakeDataStreamer.On("CloseConnection").Return(nil).Once()
	client := New()
	client.dataStream = fakeDataStreamer

	client.Start()
}

func TestStartFunctionShouldStopAndCloseTheConnectionIfReadErrorReturns(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)

	fakeDataStreamer.On("ReadByte").Return(byte(0), errors.New("cannot read")).Once()
	fakeDataStreamer.On("CloseConnection").Return(nil).Once()
	client := New()
	client.dataStream = fakeDataStreamer

	client.Start()
}
func TestStartFunctionShouldCallParseFunctionIfNoErrorWhoAmICommand(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeDataStreamer.On("ReadByte").Return(byte(0), nil)
	fakeProtocolParser.On("ParseStreamedData", mock.Anything).Return(fakeWhoAmICommand, nil)
	fakeCommandChannels.On("Add", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		assert.Equal(t, args.Get(0), fakeWhoAmICommand)
	})

	client := New()

	client.commandChannels = fakeCommandChannels
	client.dataStream = fakeDataStreamer
	client.protocolParser = fakeProtocolParser

	go client.Start()
}
func TestStartFunctionShouldCallParseFunctionIfNoErrorListClients(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeDataStreamer.On("ReadByte").Return(byte(0), nil)
	fakeProtocolParser.On("ParseStreamedData", mock.Anything).Return(fakeConnectedClientsCommand, nil)
	fakeCommandChannels.On("Add", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		assert.Equal(t, args.Get(0), fakeConnectedClientsCommand)
	})

	client := New()

	client.commandChannels = fakeCommandChannels
	client.dataStream = fakeDataStreamer
	client.protocolParser = fakeProtocolParser

	go client.Start()
}

func TestStartFunctionShouldCallParseFunctionIfNoErrorMessageFromClient(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeDataStreamer.On("ReadByte").Return(byte(0), nil)
	fakeProtocolParser.On("ParseStreamedData", mock.Anything).Return(fakeMessageFromClientCommand, nil)
	fakeCommandChannels.On("Add", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		assert.Equal(t, args.Get(0), fakeMessageFromClientCommand)
	})

	client := New()

	client.commandChannels = fakeCommandChannels
	client.dataStream = fakeDataStreamer
	client.protocolParser = fakeProtocolParser

	go client.Start()
}

func TestStartFunctionShouldCallParseFunctionIfNoErrorReturnReadErrorIfCannotParse(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)

	fakeDataStreamer.On("ReadByte").Return(byte(0), nil).Once()
	fakeProtocolParser.On("ParseStreamedData", mock.Anything).Return(nil, errors.New("error")).Once()

	client := New()

	client.dataStream = fakeDataStreamer
	client.protocolParser = fakeProtocolParser

	client.Start()
}

func TestCloseFunctionShouldReturnNoError(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)

	fakeDataStreamer.On("CloseConnection").Return(nil)
	client := New()

	client.dataStream = fakeDataStreamer

	response := client.Close()

	assert.Nil(t, response)
}

func TestCloseFunctionShouldReturnError(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeError := errors.New("cannot close")
	fakeDataStreamer.On("CloseConnection").Return(fakeError)
	client := New()

	client.dataStream = fakeDataStreamer

	response := client.Close()

	assert.Error(t, response, fakeError)
}

func TestWhoAmIFunctionShouldReturnID(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeDataStreamer.On("Write", mock.Anything).Return(0, nil)
	fakeDataStreamer.On("Flush").Return(nil)
	fakeCommandChannels.On("Get", mock.Anything).Return(fakeWhoAmICommand, nil).Run(func(args mock.Arguments) {
		assert.Equal(t, args.Get(0), protocol.CommandTypeWhoAmI)
	})

	client := &Client{
		commandChannels: fakeCommandChannels,
		dataStream:      fakeDataStreamer,
		protocolParser:  fakeProtocolParser,
	}

	response, err := client.WhoAmI()
	assert.Equal(t, response, fakeWhoAmICommand.ClientID)
	assert.Equal(t, err, nil)
}

func TestWhoAmIFunctionShouldReturnErrorIfWriteReturnError(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeWriteError := errors.New("write error")
	fakeDataStreamer.On("Write", mock.Anything).Return(0, fakeWriteError)

	client := &Client{
		commandChannels: fakeCommandChannels,
		dataStream:      fakeDataStreamer,
		protocolParser:  fakeProtocolParser,
	}

	response, err := client.WhoAmI()
	assert.Equal(t, uint64(0), response)
	assert.Equal(t, err, fakeWriteError)
}

func TestWhoAmIFunctionShouldReturnErrorIfFlushReturnError(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeWriteError := errors.New("write error")
	fakeDataStreamer.On("Write", mock.Anything).Return(0, nil)
	fakeDataStreamer.On("Flush").Return(fakeWriteError)

	client := &Client{
		commandChannels: fakeCommandChannels,
		dataStream:      fakeDataStreamer,
		protocolParser:  fakeProtocolParser,
	}

	response, err := client.WhoAmI()
	assert.Equal(t, uint64(0), response)
	assert.Equal(t, err, fakeWriteError)
}

func TestWhoAmIFunctionShouldReturnErrorIfGetReturnError(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeGetError := errors.New("get error")
	fakeDataStreamer.On("Write", mock.Anything).Return(0, nil)
	fakeDataStreamer.On("Flush").Return(nil)
	fakeCommandChannels.On("Get", mock.Anything).Return(nil, fakeGetError).Run(func(args mock.Arguments) {
		assert.Equal(t, args.Get(0), protocol.CommandTypeWhoAmI)
	})

	client := &Client{
		commandChannels: fakeCommandChannels,
		dataStream:      fakeDataStreamer,
		protocolParser:  fakeProtocolParser,
	}

	response, err := client.WhoAmI()
	assert.Equal(t, uint64(0), response)
	assert.Equal(t, err, fakeGetError)
}

func TestListClientIDsFunctionShouldReturnConnectedClients(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeDataStreamer.On("Write", mock.Anything).Return(0, nil)
	fakeDataStreamer.On("Flush").Return(nil)
	fakeCommandChannels.On("Get", mock.Anything).Return(fakeConnectedClientsCommand, nil).Run(func(args mock.Arguments) {
		assert.Equal(t, args.Get(0), protocol.CommandTypeListClients)
	})

	client := &Client{
		commandChannels: fakeCommandChannels,
		dataStream:      fakeDataStreamer,
		protocolParser:  fakeProtocolParser,
	}

	response, err := client.ListClientIDs()
	assert.Equal(t, response, fakeConnectedClientsCommand.ConnectedClients)
	assert.Equal(t, err, nil)
}

func TestListConnectedClientsFunctionShouldReturnErrorIfWriteReturnError(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeWriteError := errors.New("write error")
	fakeDataStreamer.On("Write", mock.Anything).Return(0, fakeWriteError)

	client := &Client{
		commandChannels: fakeCommandChannels,
		dataStream:      fakeDataStreamer,
		protocolParser:  fakeProtocolParser,
	}

	response, err := client.ListClientIDs()
	assert.Nil(t, response)
	assert.Equal(t, err, fakeWriteError)
}

func TestListClientsFunctionShouldReturnErrorIfFlushReturnError(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeWriteError := errors.New("write error")
	fakeDataStreamer.On("Write", mock.Anything).Return(0, nil)
	fakeDataStreamer.On("Flush").Return(fakeWriteError)

	client := &Client{
		commandChannels: fakeCommandChannels,
		dataStream:      fakeDataStreamer,
		protocolParser:  fakeProtocolParser,
	}

	response, err := client.ListClientIDs()
	assert.Nil(t, response)
	assert.Equal(t, err, fakeWriteError)
}

func TestListClientsFunctionShouldReturnErrorIfGetReturnError(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeGetError := errors.New("get error")
	fakeDataStreamer.On("Write", mock.Anything).Return(0, nil)
	fakeDataStreamer.On("Flush").Return(nil)
	fakeCommandChannels.On("Get", mock.Anything).Return(nil, fakeGetError).Run(func(args mock.Arguments) {
		assert.Equal(t, args.Get(0), protocol.CommandTypeListClients)
	})

	client := &Client{
		commandChannels: fakeCommandChannels,
		dataStream:      fakeDataStreamer,
		protocolParser:  fakeProtocolParser,
	}

	response, err := client.ListClientIDs()
	assert.Nil(t, response)
	assert.Equal(t, err, fakeGetError)
}

func TestSendMsgFunctionShouldSendMessageToServer(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeDataStreamer.On("Write", mock.Anything).Return(0, nil)
	fakeDataStreamer.On("Flush").Return(nil)

	client := &Client{
		commandChannels: fakeCommandChannels,
		dataStream:      fakeDataStreamer,
		protocolParser:  fakeProtocolParser,
	}

	err := client.SendMsg([]uint64{uint64(1)}, []byte("message"))
	assert.Equal(t, err, nil)
}

func TestSendMessageFunctionShouldReturnErrorIfWriteReturnError(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeWriteError := errors.New("write error")
	fakeDataStreamer.On("Write", mock.Anything).Return(0, fakeWriteError)

	client := &Client{
		commandChannels: fakeCommandChannels,
		dataStream:      fakeDataStreamer,
		protocolParser:  fakeProtocolParser,
	}

	err := client.SendMsg([]uint64{uint64(1)}, []byte("message"))
	assert.Equal(t, err, fakeWriteError)
}

func TestSendMessageFunctionShouldReturnErrorIfFlushReturnError(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	fakeWriteError := errors.New("write error")
	fakeDataStreamer.On("Write", mock.Anything).Return(0, nil)
	fakeDataStreamer.On("Flush").Return(fakeWriteError)

	client := &Client{
		commandChannels: fakeCommandChannels,
		dataStream:      fakeDataStreamer,
		protocolParser:  fakeProtocolParser,
	}

	err := client.SendMsg([]uint64{uint64(1)}, []byte("message"))
	assert.Equal(t, err, fakeWriteError)
}

func TestHandleIncomingMsgShouldWriteToChannel(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	client := &Client{
		commandChannels: fakeCommandChannels,
		dataStream:      fakeDataStreamer,
		protocolParser:  fakeProtocolParser,
	}
	fakeCommandChannels.On("Get", mock.Anything).Return(fakeMessageFromClientCommand, nil).Run(func(args mock.Arguments) {
		assert.Equal(t, args.Get(0), protocol.CommandTypeMessageFromClient)
	})

	incomingChan := make(chan<- protocol.MessageFromClient)
	go client.HandleIncomingMessages(incomingChan)
	time.Sleep(20 * time.Millisecond) // to ensure above routine started
}

func TestHandleIncomingMsgShouldBeRecoveredIfErrorOccur(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeCommandChannels := new(channels.MockCommandChannels)

	client := &Client{
		commandChannels: fakeCommandChannels,
		dataStream:      fakeDataStreamer,
		protocolParser:  fakeProtocolParser,
	}
	fakeCommandChannels.On("Get", mock.Anything).Return(fakeMessageFromClientCommand, errors.New("read error")).Run(func(args mock.Arguments) {
		assert.Equal(t, args.Get(0), protocol.CommandTypeMessageFromClient)
	})

	incomingChan := make(chan<- protocol.MessageFromClient)
	go client.HandleIncomingMessages(incomingChan)
	time.Sleep(20 * time.Millisecond) // to ensure above routine started
}
