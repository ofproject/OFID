package ofworldnodes

import (
	"encoding/json"
	"errors"
	"github.com/nw/ofworldnodes/db"
	//"github.com/syndtr/goleveldb/leveldb"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"os"
)

var (
	ledb        *db.Table = nil
	defaultPath           = "./db/shamirdb"
)

type RpcNode uint8
type Rpcapi struct {
	lis    *net.TCPListener
	c      *Client
	closed bool
}

type SecretRequest struct {
	Key  []byte
	Data []byte
	//Data []Shamir
}

func (n *RpcNode) Save(request *SecretRequest, result *bool) error {
	*result = true
	if request == nil {
		Logger.Error("request is invalid")
		return errors.New("request is invalid")
	}

	//return ledb.Put(request.Key, request.Data, nil)
	return ledb.PutShamir(request.Key, request.Data)
}

func (n *RpcNode) Get(key []byte, reslut *[]Shamir) error {

	var secrets []Shamir

	secretsByte, err := ledb.GetShamirSecrets(key)
	//ledb.Get(key, nil)
	if err != nil {
		return err
	}

	if jsonErr := json.Unmarshal(secretsByte, &secrets); jsonErr != nil {
		Logger.Errorf("Fail to unmarshal: ", jsonErr)
		return jsonErr
	}

	*reslut = secrets

	return nil
}

func (r *Rpcapi) StartForTCPCodecJson(ip string, client *Client) {

	rcvr := new(RpcNode)
	rpc.Register(rcvr)
	r.c = client
	r.closed = false

	tcpAddr, err := net.ResolveTCPAddr("tcp", ip)
	if err != nil {
		Logger.Error("Resolve TCPAddr failed ", ip, " err:", err)
		return
	}

	lis, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		os.Exit(400)
	}

	r.lis = lis

	dbs, err := db.NewTable(defaultPath)
	if err != nil {
		Logger.Error("Get DB Table failed: ", err)
		return
	}

	ledb = dbs
	defer dbs.Close()

	for {
		conn, err := r.lis.AcceptTCP()
		if err != nil {
			Logger.Error("RPC Service Accept Error: ", err)
			if !r.closed {
				r.c.Disconnect()
			}
			return
		}

		conn.SetNoDelay(true)

		go jsonrpc.ServeConn(conn)
	}
}

func (r *Rpcapi) Close() {
	if r.lis != nil {
		r.lis.Close()
		r.lis = nil
	}

}
