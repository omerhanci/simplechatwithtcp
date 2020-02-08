package datastream

import (
	"bufio"
	"net"
)

type TcpDataStream struct {
	conn     net.Conn
	listener net.Listener
	writer   *bufio.Writer
	reader   *bufio.Reader
}

type TcpDataStreamProducer struct {
}

func (t *TcpDataStreamProducer) Produce() IDataStreamer {
	return &TcpDataStream{}
}

func (dataStream *TcpDataStream) ReadByte() (byte, error) {
	return dataStream.reader.ReadByte()
}

func (dataStream *TcpDataStream) CloseConnection() error {
	return dataStream.conn.Close()
}

func (dataStream *TcpDataStream) CloseListener() error {
	return dataStream.listener.Close()
}

func (dataStream *TcpDataStream) Write(data []byte) (nn int, err error) {
	return dataStream.writer.Write(data)
}

func (dataStream *TcpDataStream) Flush() error {
	return dataStream.writer.Flush()
}

func (dataStream *TcpDataStream) Accept() (IDataStreamer, error) {
	conn, err := dataStream.listener.Accept()
	if err != nil {
		return nil, err
	}
	return &TcpDataStream{
		conn:   conn,
		writer: bufio.NewWriter(conn),
		reader: bufio.NewReader(conn),
	}, nil

}

func (dataStream *TcpDataStream) CreateConnection(serverAddr *net.TCPAddr) (IDataStreamer, error) {
	conn, err := net.Dial("tcp", serverAddr.String())
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	return &TcpDataStream{
		conn:   conn,
		writer: writer,
		reader: reader,
	}, nil
}

func (dataStream *TcpDataStream) CreateListener(serverAddr *net.TCPAddr) (IDataStreamer, error) {
	listener, err := net.Listen("tcp", serverAddr.String())
	if err != nil {
		return nil, err
	}
	return &TcpDataStream{
		listener: listener,
	}, nil
}
