package protocol

type IProtocolParser interface {
	ParseStreamedData(readedByte byte) (interface{}, error)
}

type IProtocolParserProducer interface {
	Produce() IProtocolParser
}
