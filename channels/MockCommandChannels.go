package channels

import (
	"github.com/Applifier/golang-backend-assignment/protocol"
	"github.com/stretchr/testify/mock"
)

type MockCommandChannels struct {
	mock.Mock
}

func (m *MockCommandChannels) Add(data interface{}) error {
	args := m.Called(data)
	return args.Error(0)
}

func (m *MockCommandChannels) Get(commandType protocol.CommandType) (interface{}, error) {
	args := m.Called(commandType)
	return args.Get(0), args.Error(1)
}
