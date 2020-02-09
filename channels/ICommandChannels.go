package channels

import "github.com/Applifier/golang-backend-assignment/protocol"

type ICommandChannels interface {
	Add(data interface{}) error
	Get(commandType protocol.CommandType) (interface{}, error)
}

type ICommandChannelsProducer interface {
	Produce() ICommandChannels
}
