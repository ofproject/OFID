package ofworldserver

import (
	//"net"
	"net/rpc"
	"net/rpc/jsonrpc"
)

type RpcNode struct {
	cli    *rpc.Client
	closed bool
	ip     string
}

func NewRpcConnect(saddr string) *RpcNode {
	rpcClient, err := jsonrpc.Dial("tcp", saddr)
	if err != nil {
		Logger.Error("Connect RPC Node failed: ", err)
		return nil
	}

	if rpcClient == nil {
		Logger.Error("RPC Node connect failed")
		return nil
	}

	rpcNode := new(RpcNode)
	rpcNode.closed = false
	rpcNode.cli = rpcClient
	rpcNode.ip = saddr

	return rpcNode
}

func (r *RpcNode) Close() {
	if !r.closed {
		return
	}

	r.cli.Close()
	r.closed = true

}

/*

addr, err := net.ResolveTCPAddr("tcp", saddr)
	if err != nil {
		Logger.Errorf("Resolve TCPAddr failed: ", err)
		return nil
	}

	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		Logger.Error("Dail rpc node fail: ", err)
		return nil
	}

	client := rpc.NewClient(conn)

*/
