package ofworldserver

import (
	"io"
	"net"
	//"net/rpc/jsonrpc"
	"strings"
	"time"
)

const sendingQueueLength = 100

type receipt chan struct{}

func (r receipt) wait() {
	<-r
}

type sendPackage struct {
	msg Message
	r   receipt
}

type Node struct {
	s         *Service
	conn      net.Conn
	Done      chan struct{}
	writeChan chan sendPackage
	closed    bool
	Id        string
	Ip        string
	Weight    int
	uname     string
	upass     string
	rpcNode   *RpcNode
}

func split(s rune) bool {
	if s == ':' {
		return true
	}

	return false
}

func (h *Node) start() {
	go h.recv()
	go h.write()
}

func (h *Node) send(msg Message) {
	p := sendPackage{msg: msg}
	select {
	case h.writeChan <- p:
	case <-time.After(3 * time.Second):
		Logger.Warning("send msg timeout wrteQueue len: ", len(h.writeChan))
	}
}

func (h *Node) sendSync(msg Message) receipt {
	p := sendPackage{
		msg: msg,
		r:   make(receipt),
	}

	h.writeChan <- p
	return p.r
}

func (h *Node) recv() {
	defer func() {
		if h.closed {
			return
		}

		if h.rpcNode != nil {
			h.rpcNode.Close()
		}

		h.closed = true
		h.conn.Close()
		h.s.stats.clientDisconnect()
		globConsistent.Remove(h)

		close(h.writeChan)
		Logger.Error("Client Close")
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

				if len(mt.ClientId) < 1 || len(mt.ClientId) > 64 {
					res = RetCodeIdentifierRejected
				}

				h.Id = mt.ClientId
				h.uname = mt.Username
				h.upass = mt.Password
				h.Ip = h.conn.RemoteAddr().String()
				h.Weight = 1

				Logger.Error("local addr: ", h.conn.LocalAddr().String(), "--remoteAddr: ", h.Ip)

				connack := &ConnAck{
					ReturnCode: res,
				}

				addr := strings.FieldsFunc(h.Ip, split)
				rpcNode := NewRpcConnect(addr[0] + ":64935")
				if rpcNode == nil {
					Logger.Error("Connect RPC Nde failed")
					return
				}

				h.rpcNode = rpcNode

				if res != RetCodeAccepted {
					Logger.Warningf("Connecting refused for %v: %v", h.conn.RemoteAddr(), ConnectionErrors[res])
					return
				}

				//stup rpc RetCodeServerUnavailable

				//end
				h.send(connack)
				globConsistent.Add(h)

				//h.rpcNode.cli.Close()

				Logger.Debugf("Client connected from %v id: %v (keepTimer: %v),", h.conn.RemoteAddr(), h.Id, mt.KeepAliveTimer)
			}
		case *PingReq:
			{
				h.send(&PingResp{})
			}
		case *Disconnect:
			return
		default:
			Logger.Warningf("recv: unknwon message type %T", mt)
			return
		}
	}
}

func (h *Node) write() {
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

		//h.s.stats.messageSend()

		if _, ok := p.msg.(*Disconnect); ok {
			Logger.Debug("send disconnect message")
			return
		}

	}

}
