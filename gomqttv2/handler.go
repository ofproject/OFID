package gomqttv2

import (
	"errors"
	"io"
	"net"
	"time"
)

var ConnectionErrors = [6]error{
	nil, // Connection Accepted (not an error)
	errors.New("Connection Refused: unacceptable protocol version"),
	errors.New("Connection Refused: identifier rejected"),
	errors.New("Connection Refused: server unavailable"),
	errors.New("Connection Refused: bad user name or password"),
	errors.New("Connection Refused: not authorized"),
}

const sendingQueueLength = 10000

type receipt chan struct{}

func (r receipt) wait() {
	<-r
}

type sendPackage struct {
	msg Message
	r   receipt
}

type Handler struct {
	s         *Service
	conn      net.Conn
	cliendID  string
	Done      chan struct{}
	writeChan chan sendPackage
	closed    bool
}

func (h *Handler) start() {
	go h.recv()
	go h.write()
}

func (h *Handler) send(msg Message) {
	p := sendPackage{msg: msg}
	select {
	case h.writeChan <- p:
	case <-time.After(3 * time.Second):
		Logger.Warning("send msg timeout wrteQueue len: ", len(h.writeChan))
	}
}

func (h *Handler) sendSync(msg Message) receipt {
	p := sendPackage{
		msg: msg,
		r:   make(receipt),
	}

	h.writeChan <- p
	return p.r
}

func (h *Handler) recv() {
	defer func() {
		if h.closed {
			return
		}

		h.closed = true
		h.conn.Close()
		h.s.stats.clientDisconnect()
		manager.delHander(h.cliendID)
		close(h.writeChan)
	}()

	for {
		h.conn.SetDeadline(time.Now().Add(30 * time.Second))
		msg, err := DecodeOneMessage(h.conn)
		if err != nil {
			if err == io.EOF {
				Logger.Warning("Client Disconnect")
				return
			}

			Logger.Error("recv message fail err: ", err)
			return
		}

		h.s.stats.messageRecv()

		if h.s.Dump {
			Logger.Notice("dump in: %T", msg)
		}

		switch mt := msg.(type) {
		case *Connect:
			{
				res := RetCodeAccepted
				if mt.ProtocolName != ProtoaolName || mt.ProtocolVersion != ProtoalVersion {
					Logger.Warningf("reject connection from ", mt.ProtocolName, " version", mt.ProtocolVersion)
					res = RetCodeUnacceptableProtocolVersion
				}

				if len(mt.ClientId) < 1 || len(mt.ClientId) > 44 {
					res = RetCodeIdentifierRejected
				}

				h.cliendID = mt.ClientId

				connack := &ConnAck{
					ReturnCode: res,
				}

				h.send(connack)

				if res != RetCodeAccepted {
					Logger.Warningf("Connecting refused for %v: %v", h.conn.RemoteAddr(), ConnectionErrors[res])
					return
				}

				manager.insterHander(h)
				go func() {
					//ok := false
					for {
						pub, ok, err := DB.get(h.cliendID)
						if pub == nil && ok && err == nil {
							return
						}

						if err != nil {
							Logger.Error("Get %s fail from connects list err :", h.cliendID, err)
							return
						}

						if pub != nil {
							if h.closed {
								return
							}
							h.send(pub)
						}
					}
				}()

				Logger.Debugf("Client connected from %v id: %v (keepTimer: %v),", h.conn.RemoteAddr(), h.cliendID, mt.KeepAliveTimer)
			}
		case *Publish:
			{
				if mt.Header.QosLevel != QosAtMostOnce {
					Logger.Warningf("recv: no support for Qos %v yet", mt.Header.QosLevel)
					return
				}

				to := manager.checkHander(string(mt.TopicName))
				if to != nil {
					//to message
					Logger.Info("to message ", mt.TopicName)
					toPublish := new(Publish)
					toPublish.Header.QosLevel = QosAtMostOnce
					toPublish.TopicName = h.cliendID
					toPublish.Payload = mt.Payload
					toPublish.MessageId = mt.MessageId

					//to.sendSync(toPublish)
					to.send(toPublish)
				}

				pubAck := &PubAck{
					MessageId: mt.MessageId,
				}

				h.send(pubAck)

				tocli := manager.checkHander(mt.TopicName)
				if tocli == nil {
					DB.save(h.cliendID, mt)
				} else {
					pub := new(Publish)
					pub.MessageId = mt.MessageId
					pub.TopicName = h.cliendID
					pub.Payload = mt.Payload
					pub.Header.QosLevel = QosAtMostOnce
					tocli.send(pub)
				}

				//tocli.send(msg)
			}
		case *PubAck:
			{
				Logger.Debugf("PubAck: %d", mt.MessageId)
			}
		case *PingReq:
			{
				h.send(&PingResp{})
			}
		case *Subscribe:
			{
				if mt.Header.QosLevel != QosAtLeastOnce {
					Logger.Warning("subscribe protocol error ", *mt)
					return
				}
				if mt.MessageId == 0 {
					Logger.Warning("Invalid MessageId in ", mt)
					return
				}

				suback := &SubAck{
					MessageId: mt.MessageId,
					TopicsQos: make([]QosLevel, len(mt.Topics)),
				}

				for i, _ := range mt.Topics {
					suback.TopicsQos[i] = QosAtMostOnce
				}
				h.send(suback)
				Logger.Debug("recv: subscribe message")
			}
		case *Unsubscribe:
			{
				Logger.Debug("recv: Unsubscribe message")
			}
		case *Disconnect:
			return
		default:
			Logger.Warningf("recv: unknwon message type %T", mt)
			return
		}
	}
}

func (h *Handler) write() {
	defer func() {
		h.conn.Close()
	}()

	for p := range h.writeChan {
		if h.s.Dump {
			Logger.Notice("dump out: %T", p.msg)
		}

		err := p.msg.Encode(h.conn)
		if p.r != nil {
			close(p.r)
		}

		if err != nil {
			Logger.Error("wite to client fail error: ", err)
			return
		}

		h.s.stats.messageSend()

		if _, ok := p.msg.(*Disconnect); ok {
			Logger.Debug("send disconnect message")
			return
		}

	}

}
