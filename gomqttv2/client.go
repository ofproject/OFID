package gomqttv2

import (
	"io"
	"net"
	"time"
)

const (
	clientQueueLength = 100
)

type Client struct {
	ClientId  string
	Dump      bool
	Done      chan struct{}
	connack   chan *ConnAck
	conn      net.Conn
	writeChan chan sendPackage
	suback    chan *SubAck
	publish   chan *dbMessage
	id        uint16
}

func NewClient(c net.Conn, clientId string) *Client {
	if len(clientId) < 2 {
		Logger.Fatal("New Client failed beacuse clientid is invaild")
		return nil
	}

	client := &Client{
		conn:      c,
		ClientId:  clientId,
		Done:      make(chan struct{}),
		connack:   make(chan *ConnAck),
		writeChan: make(chan sendPackage, clientQueueLength),
		publish:   make(chan *dbMessage, clientQueueLength),
		suback:    make(chan *SubAck),
		id:        2,
	}

	go client.recv()
	go client.write()
	return client

}

func (c *Client) nexId() uint16 {
	id := c.id
	c.id++
	return id
}

func (c *Client) Connect(name, pass string) error {
	req := &Connect{
		ProtocolName:    ProtoaolName,
		ProtocolVersion: ProtoalVersion,
		ClientId:        c.ClientId,
		CleanSession:    true,
	}

	if len(name) > 0 {
		req.UsernameFlag = true
		req.Username = name
	}

	if len(pass) > 0 {
		req.PasswordFlag = true
		req.Password = pass
	}

	c.sendSync(req)
	ack := <-c.connack
	return ConnectionErrors[ack.ReturnCode]
}

func (c *Client) RecvPuback() {

	for {
		select {
		case msg, ok := <-c.publish:
			{
				if !ok {
					return
				}
				Logger.Debug("publish msg: ", msg)
			}
		case <-time.After(5 * time.Second):
			{
				//Logger.Debug()
			}
		}
	}

}

func (c *Client) recv() {
	defer func() {
		close(c.writeChan)
		c.conn.Close()
	}()

	for {
		mq, err := DecodeOneMessage(c.conn)
		if err != nil {
			if err == io.EOF {
				Logger.Warning("Server closed connect")
				return
			}
			Logger.Debug("recv error :", err)
			return
		}

		if c.Dump {
			Logger.Infof("dump in: %T", mq)
		}

		switch mt := mq.(type) {
		case *Publish:
			{
				Logger.Debugf("publish message: %s", mt.Payload)

				msg := &dbMessage{
					From:      mt.TopicName,
					To:        c.ClientId,
					MessageId: mt.MessageId,
					Data:      mt.Payload,
				}

				c.publish <- msg

				pubAck := &PubAck{
					MessageId: mt.MessageId,
				}
				c.sendSync(pubAck)
			}
		case *PubAck:
			Logger.Debugf("recv: %T", mt)
			continue
		case *ConnAck:
			c.connack <- mt
		case *SubAck:
			c.suback <- mt
		case *Unsubscribe:
		case *Disconnect:
		default:
			Logger.Warningf("recv: got msg type %T", mt)
		}
	}
}

func (c *Client) write() {
	defer func() {
		close(c.Done)
	}()

	for {
		select {
		case job, ok := <-c.writeChan:
			{
				if !ok {
					return
				}
				if c.Dump {
					Logger.Infof("dump out: %T", job.msg)
				}

				err := job.msg.Encode(c.conn)
				if job.r != nil {
					close(job.r)
				}

				if err != nil {
					Logger.Error("write data to server failed error: ", err)
				}

				if _, ok := job.msg.(*Disconnect); ok {
					return
				}
			}
		case <-time.After(20 * time.Second):
			{
				ping := new(PingReq)
				err := ping.Encode(c.conn)
				if err != nil {
					Logger.Error("send ping fail error: ", err)
				}
			}
		}
	}
}

func (c *Client) sendSync(msg Message) {
	j := sendPackage{
		msg: msg,
		r:   make(receipt),
	}

	c.writeChan <- j
	<-j.r
	return
}

func (c *Client) Wait() {
	<-c.Done
}

func (c *Client) Disconnect() {
	c.sendSync(&Disconnect{})
	<-c.Done
}

func (c *Client) Subscribe(topics []string) *SubAck {
	if len(topics) < 0 {
		Logger.Error("subscribe topics don`t empty")
		return nil
	}

	header := Header{
		DupFlag:  false,
		Retain:   false,
		QosLevel: QosAtLeastOnce,
	}

	var tq []TopicQos
	for _, v := range topics {
		t := TopicQos{
			Topic: v,
			Qos:   QosAtMostOnce,
		}

		tq = append(tq, t)
	}

	subscribe := &Subscribe{
		Header:    header,
		MessageId: c.nexId(),
		Topics:    tq,
	}

	c.writeChan <- sendPackage{msg: subscribe}
	ack := <-c.suback
	return ack
}

func (c *Client) Publish(topic string, msg []byte) {
	if len(topic) < 0 {
		Logger.Error("publish topic don`t empty")
		return
	}

	heaer := Header{
		DupFlag:  false,
		Retain:   false,
		QosLevel: QosAtMostOnce,
	}

	publish := &Publish{
		Header:    heaer,
		TopicName: topic,
		MessageId: c.nexId(),
		Payload:   msg,
	}

	c.writeChan <- sendPackage{msg: publish}
}
