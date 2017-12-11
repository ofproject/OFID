package ofworldserver

import (
	"io"
)

func NewMessage(msgType MessageType) (msg Message, err error) {
	switch msgType {
	case CONNECT:
		msg = new(Connect)
	case CONNACK:
		msg = new(ConnAck)
	case PUBLISH:
		msg = new(Publish)
	case PUBACK:
		msg = new(PubAck)
	case PUBREC:
		msg = new(PubRec)
	case PUBREL:
		msg = new(PubRel)
	case PUBCOMP:
		msg = new(PubComp)
	case SUBSCRIBE:
		msg = new(Subscribe)
	case UNSUBACK:
		msg = new(UnsubAck)
	case SUBACK:
		msg = new(SubAck)
	case UNSUBSCRIBE:
		msg = new(Unsubscribe)
	case PINGREQ:
		msg = new(PingReq)
	case PINGRESP:
		msg = new(PingResp)
	case DISCONNECT:
		msg = new(Disconnect)
	default:
		return nil, badMsgTypeError
	}

	return
}

func DecodeOneMessage(r io.Reader) (Message, error) {
	var header Header

	msgType, length, err := header.Decode(r)
	if err != nil {
		return nil, err
	}

	//Logger.Debug("Message Type is :", msgType)

	msg, err := NewMessage(msgType)
	if err != nil {
		return nil, err
	}

	return msg, msg.Decode(r, header, length)
}
