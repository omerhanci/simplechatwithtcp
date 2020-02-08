package datastream

import (
	"net"
)

type IDataStreamer interface {
	CreateConnection(serverAddr *net.TCPAddr) (IDataStreamer, error)
	CreateListener(serverAddr *net.TCPAddr) (IDataStreamer, error)
	Accept() (IDataStreamer, error)
	CloseConnection() error
	CloseListener() error
	ReadByte() (byte, error)
	Write(data []byte) (nn int, err error)
	Flush() error
}

type IDataStreamerProducer interface {
	Produce() IDataStreamer
}
