package server

import (
	"errors"
	"io"
	"net"
	"sync"
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

func TestNewMethodShouldCreateNewServerInstance(t *testing.T) {
	server := New()
	assert.NotNil(t, server)
	assert.NotNil(t, server.commandChannels)
}

func TestStartFunctionShouldReturnNoErrorIfListenerIsCreated(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeCommandChannels := new(channels.MockCommandChannels)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeProtocolParserProducer := new(protocol.MockProtocolParserProducer)
	server := &Server{
		commandChannels:        fakeCommandChannels,
		dataStreamer:           fakeDataStreamer,
		protocolParserProducer: fakeProtocolParserProducer,
		clientMutex:            &sync.Mutex{}}

	fakeDataStreamer.On("CreateListener", mock.Anything).Return(fakeDataStreamer, nil).Once()
	fakeDataStreamer.On("Accept").Return(fakeDataStreamer, nil)
	fakeProtocolParserProducer.On("Produce").Return(fakeProtocolParser).Maybe()
	fakeDataStreamer.On("ReadByte", mock.Anything).Return(byte(0), nil).Maybe()
	fakeDataStreamer.On("CloseConnection").Return(nil).Maybe()
	fakeProtocolParser.On("ParseStreamedData", mock.Anything).Return(nil, nil).Maybe()

	response := server.Start(&fakeAddress)
	assert.Nil(t, response)
}

func TestStartFunctionShouldReturnErrorIfListenerCannotCreated(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeCommandChannels := new(channels.MockCommandChannels)
	fakeProtocolParserProducer := new(protocol.MockProtocolParserProducer)
	fakeProtocolParser := new(protocol.MockProtocolParser)

	server := &Server{
		commandChannels:        fakeCommandChannels,
		dataStreamer:           fakeDataStreamer,
		protocolParserProducer: fakeProtocolParserProducer,
		clientMutex:            &sync.Mutex{}}

	fakeListenerError := errors.New("cannot create listener")
	fakeDataStreamer.On("CreateListener", mock.Anything).Return(fakeDataStreamer, fakeListenerError).Once()
	fakeDataStreamer.On("Accept").Return(fakeDataStreamer, nil)
	fakeProtocolParserProducer.On("Produce").Return(fakeProtocolParser).Maybe()
	fakeDataStreamer.On("ReadByte", mock.Anything).Return(byte(0), nil).Maybe()
	fakeDataStreamer.On("CloseConnection").Return(nil).Maybe()
	fakeProtocolParser.On("ParseStreamedData", mock.Anything).Return(nil, nil).Maybe()

	response := server.Start(&fakeAddress)
	assert.Error(t, response)
}

func TestAcceptFunctionShouldCreateClientObjectAndSetClientID(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeCommandChannels := new(channels.MockCommandChannels)
	fakeProtocolParserProducer := new(protocol.MockProtocolParserProducer)
	server := &Server{
		commandChannels:        fakeCommandChannels,
		dataStreamer:           fakeDataStreamer,
		protocolParserProducer: fakeProtocolParserProducer,
		clientMutex:            &sync.Mutex{}}

	response := server.accept(fakeDataStreamer)
	assert.Equal(t, response.id, uint64(1))
	assert.Equal(t, response.dataStreamer, fakeDataStreamer)
}
func TestServeFunctionShouldCloseConnectionIfEOFRead(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeCommandChannels := new(channels.MockCommandChannels)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeProtocolParserProducer := new(protocol.MockProtocolParserProducer)
	server := &Server{
		commandChannels:        fakeCommandChannels,
		dataStreamer:           fakeDataStreamer,
		protocolParser:         fakeProtocolParser,
		protocolParserProducer: fakeProtocolParserProducer,
		clientMutex:            &sync.Mutex{}}

	fakeClient := &client{
		dataStreamer: fakeDataStreamer,
		id:           uint64(1),
	}

	fakeDataStreamer.On("ReadByte", mock.Anything).Return(byte(0), io.EOF).Once()
	fakeDataStreamer.On("CloseConnection").Return(nil).Once()
	fakeProtocolParserProducer.On("Produce").Return(fakeProtocolParser).Once()

	go server.serve(fakeClient)
	time.Sleep(20 * time.Millisecond) // to ensure above routine started
}

func TestServeFunctionShouldCloseConnectionIfReadError(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeCommandChannels := new(channels.MockCommandChannels)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeProtocolParserProducer := new(protocol.MockProtocolParserProducer)
	server := &Server{
		commandChannels:        fakeCommandChannels,
		dataStreamer:           fakeDataStreamer,
		protocolParser:         fakeProtocolParser,
		protocolParserProducer: fakeProtocolParserProducer,
		clientMutex:            &sync.Mutex{}}

	fakeClient := &client{
		dataStreamer: fakeDataStreamer,
		id:           uint64(1),
	}

	fakeDataStreamer.On("ReadByte", mock.Anything).Return(byte(0), errors.New("read error")).Once()
	fakeDataStreamer.On("CloseConnection").Return(nil).Once()
	fakeProtocolParserProducer.On("Produce").Return(fakeProtocolParser).Once()

	go server.serve(fakeClient)
	time.Sleep(20 * time.Millisecond) // to ensure above routine started
}

func TestServeFunctionShouldCloseConnectionIfParseError(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeCommandChannels := new(channels.MockCommandChannels)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeProtocolParserProducer := new(protocol.MockProtocolParserProducer)
	server := &Server{
		commandChannels:        fakeCommandChannels,
		dataStreamer:           fakeDataStreamer,
		protocolParser:         fakeProtocolParser,
		protocolParserProducer: fakeProtocolParserProducer,
		clientMutex:            &sync.Mutex{}}

	fakeClient := &client{
		dataStreamer: fakeDataStreamer,
		id:           uint64(1),
	}

	fakeDataStreamer.On("ReadByte", mock.Anything).Return(byte(0), nil).Once()
	fakeDataStreamer.On("CloseConnection").Return(nil).Once()
	fakeProtocolParser.On("ParseStreamedData", mock.Anything).Return(fakeWhoAmICommand, errors.New("parse error")).Once()
	fakeProtocolParserProducer.On("Produce").Return(fakeProtocolParser).Once()

	go server.serve(fakeClient)
	time.Sleep(20 * time.Millisecond) // to ensure above routine started
}

func TestServeFunctionShouldSendWhoAmICommandIfReceiveWhoAmICommand(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeCommandChannels := new(channels.MockCommandChannels)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeProtocolParserProducer := new(protocol.MockProtocolParserProducer)
	server := &Server{
		commandChannels:        fakeCommandChannels,
		dataStreamer:           fakeDataStreamer,
		protocolParser:         fakeProtocolParser,
		protocolParserProducer: fakeProtocolParserProducer,
		clientMutex:            &sync.Mutex{}}

	fakeClient := &client{
		dataStreamer: fakeDataStreamer,
		id:           uint64(1),
	}

	fakeDataStreamer.On("ReadByte", mock.Anything).Return(byte(0), nil)
	fakeProtocolParser.On("ParseStreamedData", mock.Anything).Return(fakeWhoAmICommand, nil)
	fakeProtocolParserProducer.On("Produce").Return(fakeProtocolParser).Once()

	fakeDataStreamer.On("Write", mock.Anything).Return(0, nil).Run(func(args mock.Arguments) {
		assert.Equal(t, args.Get(0).([]byte)[0], byte(protocol.CommandTypeWhoAmI))
	})
	fakeDataStreamer.On("Flush", mock.Anything).Return(nil)

	go server.serve(fakeClient)
	time.Sleep(20 * time.Millisecond) // to ensure above routine started
}

func TestServeFunctionShouldSendConnectedClientsIfReceiveListClientsCommand(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeCommandChannels := new(channels.MockCommandChannels)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeProtocolParserProducer := new(protocol.MockProtocolParserProducer)
	server := &Server{
		commandChannels:        fakeCommandChannels,
		dataStreamer:           fakeDataStreamer,
		protocolParser:         fakeProtocolParser,
		protocolParserProducer: fakeProtocolParserProducer,
		clientMutex:            &sync.Mutex{}}

	fakeClient := &client{
		dataStreamer: fakeDataStreamer,
		id:           uint64(1),
	}

	fakeDataStreamer.On("ReadByte", mock.Anything).Return(byte(0), nil)
	fakeProtocolParser.On("ParseStreamedData", mock.Anything).Return(fakeConnectedClientsCommand, nil)
	fakeProtocolParserProducer.On("Produce").Return(fakeProtocolParser).Once()

	fakeDataStreamer.On("Write", mock.Anything).Return(0, nil).Run(func(args mock.Arguments) {
		assert.Equal(t, args.Get(0).([]byte)[0], byte(protocol.CommandTypeListClients))
	})
	fakeDataStreamer.On("Flush", mock.Anything).Return(nil)

	go server.serve(fakeClient)
	time.Sleep(20 * time.Millisecond) // to ensure above routine started
}

func TestServeFunctionShouldSendMessageIfReceiveSendMessageCommand(t *testing.T) {

	fakeDataStreamer := new(datastream.MockTcpDataStream)
	fakeCommandChannels := new(channels.MockCommandChannels)
	fakeProtocolParser := new(protocol.MockProtocolParser)
	fakeProtocolParserProducer := new(protocol.MockProtocolParserProducer)
	fakeClient := &client{
		dataStreamer: fakeDataStreamer,
		id:           uint64(1),
	}
	server := &Server{
		commandChannels:        fakeCommandChannels,
		dataStreamer:           fakeDataStreamer,
		protocolParser:         fakeProtocolParser,
		protocolParserProducer: fakeProtocolParserProducer,
		clients:                []*client{fakeClient},
		clientIDs:              []uint64{fakeClient.id},
		clientMutex:            &sync.Mutex{}}

	fakeDataStreamer.On("ReadByte", mock.Anything).Return(byte(0), nil)
	fakeProtocolParser.On("ParseStreamedData", mock.Anything).Return(fakeSendMsgCommand, nil)
	fakeProtocolParserProducer.On("Produce").Return(fakeProtocolParser).Once()

	fakeDataStreamer.On("Write", mock.Anything).Return(0, nil).Run(func(args mock.Arguments) {
		assert.Equal(t, args.Get(0).([]byte)[0], byte(protocol.CommandTypeMessageFromClient))
	})
	fakeDataStreamer.On("Flush", mock.Anything).Return(nil)

	go server.serve(fakeClient)
	time.Sleep(20 * time.Millisecond) // to ensure above routine started
}
