package ofworldnodes

import (
	"errors"
)

type MessageType uint8

const (
	CONNECT     = MessageType(iota + 1) //客户端请求连接服务端
	CONNACK                             //连接报文确认 服务端到客户端
	PUBLISH                             //发布消息，双向
	PUBACK                              //QoS1消息发布收到确认，双向
	PUBREC                              //发布收到，保证交付第一步，双向
	PUBREL                              //发布释放，保证交付第二步，双向
	PUBCOMP                             //QoS2消息发布完成，保证交付第三步，双向
	SUBSCRIBE                           //客户type端订阅请求，客户端到服务端
	SUBACK                              //订阅请报文确认，服务端到客户端
	UNSUBSCRIBE                         //客户端取消订阅请求，客户端到服务端
	UNSUBACK                            //取消订阅报文确认，服务端到客户端
	PINGREQ                             //心跳请求，客户端到服务端
	PINGRESP                            //心跳响应，服务端到客户端
	DISCONNECT                          //客户端断开连接，客户端到服务端

	MSGTYPEINVALID
)

func (m MessageType) IsValid() bool {
	return m >= CONNECT && m < MSGTYPEINVALID
}

const (
	ProtoaolName     string = "MQTT"
	ProtoalVersion   uint8  = 0x04
	ProtoalMaxLength int    = 268435455
)

var (
	badMsgTypeError        = errors.New("mqtt: message type is invalid")
	badQosError            = errors.New("mqtt: QoS is invalid")
	badWillQosError        = errors.New("mqtt: will QoS is invalid")
	badLengthEncodingError = errors.New("mqtt: remaining length field exceeded maximum of 4 bytes")
	badReturnCodeError     = errors.New("mqtt: is invalid")
	dataExceedsPacketError = errors.New("mqtt: data exceeds packet length")
	msgTooLongError        = errors.New("mqtt: message is too long")
)

const (
	QosAtMostOnce = QosLevel(iota)
	QosAtLeastOnce
	QosExactlyOnce

	qosFirstInvalid
)

type QosLevel uint8

func (qos QosLevel) IsValid() bool {
	return qos < qosFirstInvalid
}

func (qos QosLevel) HasId() bool {
	return qos == QosAtLeastOnce || qos == QosExactlyOnce
}

type ReturnCode uint8

const (
	RetCodeAccepted = ReturnCode(iota)
	RetCodeUnacceptableProtocolVersion
	RetCodeIdentifierRejected
	RetCodeServerUnavailable
	RetCodeBadUsernameOrPassword
	RetCodeNotAuthorized

	retCodeFirstInvalid
)

func (rc ReturnCode) IsValid() bool {
	return rc >= RetCodeAccepted && rc < retCodeFirstInvalid
}

var ConnectionErrors = [6]error{
	nil, // Connection Accepted (not an error)
	errors.New("Connection Refused: unacceptable protocol version"),
	errors.New("Connection Refused: identifier rejected"),
	errors.New("Connection Refused: server unavailable"),
	errors.New("Connection Refused: bad user name or password"),
	errors.New("Connection Refused: not authorized"),
}
