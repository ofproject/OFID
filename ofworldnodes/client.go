package ofworldnodes

import (
	"errors"
	"io"
	"net"
	"sync/atomic"
	"time"
)

const (
	clientQueueLength = 100
)

//test
var (
	count int64 = 0
)

func countAdd() {
	atomic.AddInt64(&count, 1)
}

//end

type receipt chan struct{}

func (r receipt) wait() {
	<-r
}

type sendPackage struct {
	msg Message
	r   receipt
}

type dbMessage struct {
	Key       string `json:"Key"`
	DB        string `json:"DB"`
	MessageId string `json:"MessageID"`
	Data      []byte `json:"Body"`
}

type Client struct {
	ClientId  string
	Dump      bool
	Done      chan struct{}
	connack   chan *ConnAck
	conn      net.Conn
	writeChan chan sendPackage
	id        uint32
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
		id:        1,
	}

	go client.recv()
	go client.write()
	return client

}

func (c *Client) nexId() uint32 {
	if c.id == 0xFFFFFFFF {
		c.id = 1
	}

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
	ack, ok := <-c.connack
	if !ok {
		return errors.New("server close")
	}

	return ConnectionErrors[ack.ReturnCode]
}

func (c *Client) recv() {
	defer func() {
		close(c.Done)
		close(c.writeChan)
		close(c.connack)
		c.conn.Close()
		//Logger.Error("...")
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
		case *ConnAck:
			c.connack <- mt
		case *Disconnect:
		case *PingResp:
		default:
			Logger.Warningf("recv: got msg type %T", mt)
		}
	}
}

func (c *Client) write() {
	defer func() {
		//close(c.Done)
		c.conn.Close()
		//Logger.Error("...")
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
