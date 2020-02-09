package datastream

import (
	"net"

	"github.com/stretchr/testify/mock"
)

type MockTcpDataStream struct {
	mock.Mock
}

func (m *MockTcpDataStream) CreateConnection(serverAddr *net.TCPAddr) (IDataStreamer, error) {
	args := m.Called(serverAddr)
	return args.Get(0).(IDataStreamer), args.Error(1)
}

func (m *MockTcpDataStream) CreateListener(serverAddr *net.TCPAddr) (IDataStreamer, error) {
	args := m.Called(serverAddr)
	return args.Get(0).(IDataStreamer), args.Error(1)
}

func (m *MockTcpDataStream) Accept() (IDataStreamer, error) {
	args := m.Called()
	return args.Get(0).(IDataStreamer), args.Error(1)
}

func (m *MockTcpDataStream) CloseConnection() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTcpDataStream) CloseListener() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockTcpDataStream) ReadByte() (byte, error) {
	args := m.Called()
	return args.Get(0).(byte), args.Error(1)
}

func (m *MockTcpDataStream) Write(data []byte) (int, error) {
	args := m.Called(data)
	return args.Int(0), args.Error(1)
}

func (m *MockTcpDataStream) Flush() error {
	args := m.Called()
	return args.Error(0)
}
