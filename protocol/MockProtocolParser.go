package protocol

import (
	"github.com/stretchr/testify/mock"
)

type MockProtocolParser struct {
	mock.Mock
}

type MockProtocolParserProducer struct {
	mock.Mock
}

func (m *MockProtocolParserProducer) Produce() IProtocolParser {
	args := m.Called()
	return args.Get(0).(IProtocolParser)
}

func (m *MockProtocolParser) ParseStreamedData(readedByte byte) (interface{}, error) {
	args := m.Called(readedByte)
	return args.Get(0), args.Error(1)
}
