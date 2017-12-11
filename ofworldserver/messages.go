package ofworldserver

import (
	"bytes"
	"io"
)

const (
	MaxLength = (1 << (4 * 7)) - 1
)

type Header struct {
	DupFlag, Retain bool
	QosLevel        QosLevel
}

func (h *Header) Encode(w io.Writer, msgType MessageType, remainingLength int32) error {
	buf := new(bytes.Buffer)
	err := h.encodeInto(buf, msgType, remainingLength)
	if err != nil {
		return err
	}
	_, err = w.Write(buf.Bytes())
	return err
}

func (h *Header) encodeInto(buf *bytes.Buffer, msgType MessageType, remainingLength int32) error {
	if !h.QosLevel.IsValid() {
		return badQosError
	}
	if !msgType.IsValid() {
		return badMsgTypeError
	}

	val := byte(msgType) << 4
	buf.WriteByte(val)
	encodeLength(remainingLength, buf)
	return nil
}

func (h *Header) Decode(r io.Reader) (msgType MessageType, remainingLength int32, err error) {
	defer func() {
		err = recoverError(err, recover())
	}()

	var buf [1]byte

	if _, err = io.ReadFull(r, buf[:]); err != nil {
		return
	}

	byte1 := buf[0]
	msgType = MessageType(byte1 & 0xF0 >> 4)

	*h = Header{
		DupFlag:  byte1&0x08 > 0,
		QosLevel: QosLevel(byte1 & 0x06 >> 1),
		Retain:   byte1&0x01 > 0,
	}

	remainingLength = decodeLength(r)

	return
}

type Message interface {
	Encode(w io.Writer) error
	Decode(r io.Reader, header Header, packetRemaining int32) error
}

func writeMessage(w io.Writer, msgType MessageType, header *Header, msgBuf *bytes.Buffer) error {
	totalLength := int64(len(msgBuf.Bytes()))
	if totalLength > MaxLength {
		return msgTooLongError
	}

	buf := new(bytes.Buffer)
	err := header.encodeInto(buf, msgType, int32(totalLength))
	if err != nil {
		return err
	}

	buf.Write(msgBuf.Bytes())
	_, err = w.Write(buf.Bytes())

	return err
}

type Connect struct {
	Header
	ProtocolName               string
	ProtocolVersion            uint8
	WillRetain                 bool
	WillFlag                   bool
	CleanSession               bool
	WillQos                    QosLevel
	KeepAliveTimer             uint16
	ClientId                   string
	WillTopic, WillMessage     string
	UsernameFlag, PasswordFlag bool
	Username, Password         string
}

func (msg *Connect) Encode(w io.Writer) (err error) {
	if !msg.WillQos.IsValid() {
		return badWillQosError
	}

	buf := new(bytes.Buffer)

	flags := boolToByte(msg.UsernameFlag) << 7
	flags |= boolToByte(msg.PasswordFlag) << 6
	flags |= boolToByte(msg.WillRetain) << 5
	flags |= byte(msg.WillQos) << 3
	flags |= boolToByte(msg.WillFlag) << 2
	flags |= boolToByte(msg.CleanSession) << 1

	setString(msg.ProtocolName, buf)
	setUint8(msg.ProtocolVersion, buf)
	buf.WriteByte(flags)
	setUint16(msg.KeepAliveTimer, buf)
	setString(msg.ClientId, buf)
	if msg.WillFlag {
		setString(msg.WillTopic, buf)
		setString(msg.WillMessage, buf)
	}
	if msg.UsernameFlag {
		setString(msg.Username, buf)
	}
	if msg.PasswordFlag {
		setString(msg.Password, buf)
	}

	return writeMessage(w, CONNECT, &msg.Header, buf)
}

func (msg *Connect) Decode(r io.Reader, header Header, packetRemaining int32) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()

	msg.Header = header

	protocolName := getString(r, &packetRemaining)
	protocolVersion := getUint8(r, &packetRemaining)
	flags := getUint8(r, &packetRemaining)
	keepAliveTimer := getUint16(r, &packetRemaining)
	clientId := getString(r, &packetRemaining)

	*msg = Connect{
		ProtocolName:    protocolName,
		ProtocolVersion: protocolVersion,
		UsernameFlag:    flags&0x80 > 0,
		PasswordFlag:    flags&0x40 > 0,
		WillRetain:      flags&0x20 > 0,
		WillQos:         QosLevel(flags & 0x18 >> 3),
		WillFlag:        flags&0x04 > 0,
		CleanSession:    flags&0x02 > 0,
		KeepAliveTimer:  keepAliveTimer,
		ClientId:        clientId,
	}

	if msg.WillFlag {
		msg.WillTopic = getString(r, &packetRemaining)
		msg.WillMessage = getString(r, &packetRemaining)
	}

	if msg.UsernameFlag {
		msg.Username = getString(r, &packetRemaining)
	}

	if msg.PasswordFlag {
		msg.Password = getString(r, &packetRemaining)
	}

	if packetRemaining != 0 {
		return msgTooLongError
	}

	return nil
}

type ConnAck struct {
	Header
	ReturnCode ReturnCode
}

func (msg *ConnAck) Encode(w io.Writer) (err error) {
	buf := new(bytes.Buffer)

	buf.WriteByte(byte(0)) // Reserved byte.
	setUint8(uint8(msg.ReturnCode), buf)

	return writeMessage(w, CONNACK, &msg.Header, buf)
}

func (msg *ConnAck) Decode(r io.Reader, header Header, packetRemaining int32) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()

	msg.Header = header

	getUint8(r, &packetRemaining)
	msg.ReturnCode = ReturnCode(getUint8(r, &packetRemaining))
	if !msg.ReturnCode.IsValid() {
		return badReturnCodeError
	}

	if packetRemaining != 0 {
		return msgTooLongError
	}

	return nil
}

type Publish struct {
	Header
	TopicName string
	//MessageId uint16
	MessageId string
	Payload   []byte
}

func (msg *Publish) Encode(w io.Writer) (err error) {
	buf := new(bytes.Buffer)

	setString(msg.TopicName, buf)
	if msg.Header.QosLevel.HasId() {
		//setUint16(msg.MessageId, buf)
		setString(msg.MessageId, buf)
	}

	buf.Write(msg.Payload)

	return writeMessage(w, PUBLISH, &msg.Header, buf)
}

func (msg *Publish) Decode(r io.Reader, header Header, packetRemaining int32) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()

	msg.Header = header

	msg.TopicName = getString(r, &packetRemaining)
	if msg.Header.QosLevel.HasId() {
		//msg.MessageId = getUint16(r, &packetRemaining)
		msg.MessageId = getString(r, &packetRemaining)
	}

	payloadReader := &io.LimitedReader{r, int64(packetRemaining)}
	msg.Payload = make([]byte, int(packetRemaining))

	_, err = io.ReadFull(payloadReader, msg.Payload)

	return err
}

type PubAck struct {
	Header
	MessageId string
}

func encodeAckCommon(w io.Writer, header *Header, messageId string, msgType MessageType) error {
	buf := new(bytes.Buffer)
	//setUint16(messageId, buf)
	setString(messageId, buf)
	return writeMessage(w, msgType, header, buf)
}

func decodeAckCommon(r io.Reader, packetRemaining int32, messageId *string) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()

	*messageId = getString(r, &packetRemaining)

	if packetRemaining != 0 {
		return msgTooLongError
	}

	return nil
}

func (msg *PubAck) Encode(w io.Writer) error {
	return encodeAckCommon(w, &msg.Header, msg.MessageId, PUBACK)
}

func (msg *PubAck) Decode(r io.Reader, header Header, packetRemaining int32) (err error) {
	msg.Header = header
	return decodeAckCommon(r, packetRemaining, &msg.MessageId)
}

type PubRec struct {
	Header
	MessageId string
}

func (msg *PubRec) Encode(w io.Writer) error {
	return encodeAckCommon(w, &msg.Header, msg.MessageId, PUBREC)
}

func (msg *PubRec) Decode(r io.Reader, header Header, packetRemaining int32) (err error) {
	msg.Header = header
	return decodeAckCommon(r, packetRemaining, &msg.MessageId)
}

type PubRel struct {
	Header
	MessageId string
}

func (msg *PubRel) Encode(w io.Writer) error {
	return encodeAckCommon(w, &msg.Header, msg.MessageId, PUBREL)
}

func (msg *PubRel) Decode(r io.Reader, header Header, packetRemaining int32) (err error) {
	msg.Header = header
	return decodeAckCommon(r, packetRemaining, &msg.MessageId)
}

type PubComp struct {
	Header
	MessageId string
}

func (msg *PubComp) Encode(w io.Writer) error {
	return encodeAckCommon(w, &msg.Header, msg.MessageId, PUBCOMP)
}

func (msg *PubComp) Decode(r io.Reader, header Header, packetRemaining int32) (err error) {
	msg.Header = header
	return decodeAckCommon(r, packetRemaining, &msg.MessageId)
}

type Subscribe struct {
	Header
	MessageId string
	Topics    []TopicQos
}

type TopicQos struct {
	Topic string
	Qos   QosLevel
}

func (msg *Subscribe) Encode(w io.Writer) (err error) {
	buf := new(bytes.Buffer)
	if msg.Header.QosLevel.HasId() {
		//setUint16(msg.MessageId, buf)
		setString(msg.MessageId, buf)
	}
	for _, topicSub := range msg.Topics {
		setString(topicSub.Topic, buf)
		setUint8(uint8(topicSub.Qos), buf)
	}

	return writeMessage(w, SUBSCRIBE, &msg.Header, buf)
}

func (msg *Subscribe) Decode(r io.Reader, header Header, packetRemaining int32) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()

	msg.Header = header

	if msg.Header.QosLevel.HasId() {
		//msg.MessageId = getUint16(r, &packetRemaining)
		msg.MessageId = getString(r, &packetRemaining)
	}
	var topics []TopicQos
	for packetRemaining > 0 {
		topics = append(topics, TopicQos{
			Topic: getString(r, &packetRemaining),
			Qos:   QosLevel(getUint8(r, &packetRemaining)),
		})
	}
	msg.Topics = topics

	return nil
}

type SubAck struct {
	Header
	MessageId string
	TopicsQos []QosLevel
}

func (msg *SubAck) Encode(w io.Writer) (err error) {
	buf := new(bytes.Buffer)
	//setUint16(msg.MessageId, buf)
	setString(msg.MessageId, buf)
	for i := 0; i < len(msg.TopicsQos); i += 1 {
		setUint8(uint8(msg.TopicsQos[i]), buf)
	}

	return writeMessage(w, SUBACK, &msg.Header, buf)
}

func (msg *SubAck) Decode(r io.Reader, header Header, packetRemaining int32) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()

	msg.Header = header

	//msg.MessageId = getUint16(r, &packetRemaining)
	msg.MessageId = getString(r, &packetRemaining)
	topicsQos := make([]QosLevel, 0)
	for packetRemaining > 0 {
		grantedQos := QosLevel(getUint8(r, &packetRemaining) & 0x03)
		topicsQos = append(topicsQos, grantedQos)
	}
	msg.TopicsQos = topicsQos

	return nil
}

type Unsubscribe struct {
	Header
	MessageId string
	Topics    []string
}

func (msg *Unsubscribe) Encode(w io.Writer) (err error) {
	buf := new(bytes.Buffer)
	if msg.Header.QosLevel.HasId() {
		//setUint16(msg.MessageId, buf)
		setString(msg.MessageId, buf)
	}
	for _, topic := range msg.Topics {
		setString(topic, buf)
	}

	return writeMessage(w, UNSUBSCRIBE, &msg.Header, buf)
}

func (msg *Unsubscribe) Decode(r io.Reader, header Header, packetRemaining int32) (err error) {
	defer func() {
		err = recoverError(err, recover())
	}()

	msg.Header = header

	if qos := msg.Header.QosLevel; qos == 1 || qos == 2 {
		//msg.MessageId = getUint16(r, &packetRemaining)
		msg.MessageId = getString(r, &packetRemaining)
	}
	topics := make([]string, 0)
	for packetRemaining > 0 {
		topics = append(topics, getString(r, &packetRemaining))
	}
	msg.Topics = topics

	return nil
}

type UnsubAck struct {
	Header
	MessageId string
}

func (msg *UnsubAck) Encode(w io.Writer) error {
	return encodeAckCommon(w, &msg.Header, msg.MessageId, UNSUBACK)
}

func (msg *UnsubAck) Decode(r io.Reader, header Header, packetRemaining int32) (err error) {
	msg.Header = header
	return decodeAckCommon(r, packetRemaining, &msg.MessageId)
}

type PingReq struct {
	Header
}

func (msg *PingReq) Encode(w io.Writer) error {
	return msg.Header.Encode(w, PINGREQ, 0)
}

func (msg *PingReq) Decode(r io.Reader, header Header, packetRemaining int32) error {
	if packetRemaining != 0 {
		return msgTooLongError
	}
	return nil
}

// PingResp represents an MQTT PINGRESP message.
type PingResp struct {
	Header
}

func (msg *PingResp) Encode(w io.Writer) error {
	return msg.Header.Encode(w, PINGRESP, 0)
}

func (msg *PingResp) Decode(r io.Reader, header Header, packetRemaining int32) error {
	if packetRemaining != 0 {
		return msgTooLongError
	}
	return nil
}

type Disconnect struct {
	Header
}

func (msg *Disconnect) Encode(w io.Writer) error {
	return msg.Header.Encode(w, DISCONNECT, 0)
}

func (msg *Disconnect) Decode(r io.Reader, header Header, packetRemaining int32) error {
	if packetRemaining != 0 {
		return msgTooLongError
	}
	return nil
}
