package ofworldserver

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"github.com/nw/ofworldserver/shamir"
	"github.com/nw/ofworldserver/utils"
	"io"
	"net/http"
	"net/rpc"
	"strconv"
	"sync"
)

type userInfo struct {
	Addr []byte `json:"addr"`
	Hash []byte `json:"hash"`
	Data []byte `json:"data"`
}

type ResponseInfo struct {
	Addr       []byte   `json:"addr"`
	Hash       []byte   `json:"Hash"`
	NodeNumber []uint32 `json:"NodeNumber"`
}

type SecretRequest struct {
	Key []byte
	//Data []shamir.Secret
	Data []byte
}

func saveToNode(req bytes.Buffer) (jsonRes *bytes.Buffer, s int, err error) {
	info := new(userInfo)
	err = json.Unmarshal(req.Bytes(), info)
	var wg sync.WaitGroup
	if err != nil {
		Logger.Debug("Unmarshal Fail: ", err)
		return nil, http.StatusBadRequest, err
	}

	eachLength := 20

	splitByte := utils.ByteSplit(req.Bytes(), eachLength)

	result, err := shamir.GenerateShamirMul(3, 5, splitByte)
	if err != nil {
		Logger.Error("Generate shamir failed: ", err)
		return nil, http.StatusBadRequest, err
	}

	var numbers []uint32
	lock := &sync.Mutex{}

	for _, v := range result {
		wg.Add(1)
		go func(val []shamir.Secret) {
			var value []shamir.Secret = val
			defer wg.Done()
			jsonSecret, err := json.Marshal(value)

			if err != nil {
				Logger.Debug("Marshal Fail")
				return
			}
			node, number := globConsistent.Get(string(jsonSecret))
			if node == nil {
				Logger.Debug("Get Node Fail")
				return
			}

			var reply bool = false
			var bufKey bytes.Buffer
			bufKey.Write(info.Hash)
			bufKey.Write([]byte(":"))
			bufKey.Write(info.Addr)
			bufKey.Write([]byte(":"))
			bufKey.WriteString(strconv.FormatUint(uint64(number), 10))

			h := sha1.New()
			io.WriteString(h, bufKey.String())

			request := &SecretRequest{
				Key:  h.Sum(nil),
				Data: jsonSecret,
			}

			err = node.rpcNode.cli.Call("RpcNode.Save", request, &reply)
			if err != nil {
				if err == rpc.ErrShutdown {
					node.closed = true
					node.conn.Close()
				}

				return
			}

			lock.Lock()
			numbers = append(numbers, number)
			lock.Unlock()

		}(v)

	}

	wg.Wait()
	if len(numbers) != 5 {
		Logger.Debug("number: ", numbers)
		return nil, http.StatusBadRequest, errors.New("Node save data failed")
	}

	res := ResponseInfo{
		Addr:       info.Addr,
		Hash:       info.Hash,
		NodeNumber: numbers,
	}

	jsonRess, err := json.Marshal(res)
	if err != nil {
		Logger.Error("Marshal Response info failed: ", err)
		return nil, http.StatusBadRequest, err
	}

	jsonRes = new(bytes.Buffer)
	jsonRes.Write(jsonRess)

	return jsonRes, http.StatusOK, nil
}

func GetFromNode(req bytes.Buffer) (jsonRes *bytes.Buffer, s int, err error) {
	info := new(ResponseInfo)
	err = json.Unmarshal(req.Bytes(), info)
	if err != nil {
		Logger.Debug("Unmarshal Fail: ", err)
		return nil, http.StatusBadRequest, err
	}
	var wg sync.WaitGroup

	if len(info.NodeNumber) < 3 {
		Logger.Error("needkey < 3")
		return nil, http.StatusBadRequest, errors.New("number len error")
	}

	//var result []SecretRequest
	var secrets [][]shamir.Secret
	lock := &sync.Mutex{}

	for _, v := range info.NodeNumber {
		wg.Add(1)
		go func(number uint32) {
			var key uint32 = number
			defer wg.Done()
			node := globConsistent.GetFromKey(key)
			if node == nil {
				Logger.Error("get node failed")
				return
			}

			var reply []shamir.Secret

			var bufKey bytes.Buffer
			bufKey.Write(info.Hash)
			bufKey.Write([]byte(":"))
			bufKey.Write(info.Addr)
			bufKey.Write([]byte(":"))
			bufKey.WriteString(strconv.FormatUint(uint64(key), 10))

			h := sha1.New()
			io.WriteString(h, bufKey.String())

			err = node.rpcNode.cli.Call("RpcNode.Get", h.Sum(nil), &reply)
			if err != nil {
				Logger.Error("Get Failed:", err)
				if err == rpc.ErrShutdown {
					node.closed = true
					node.conn.Close()
				}

				return
			}

			lock.Lock()
			secrets = append(secrets, reply)
			lock.Unlock()

		}(v)
	}

	wg.Wait()
	if len(secrets) < 3 {
		Logger.Error("Get data not content ", len(secrets))
		return nil, http.StatusBadRequest, errors.New("Get data not content")
	}

	clearTextInt := shamir.RecoverShamirSecretsMul(secrets)
	var clearTextByte []byte
	for _, a := range clearTextInt {
		clearTextByte = append(clearTextByte, a.Bytes()...)
	}

	var userinfo = new(userInfo)
	if err = json.Unmarshal(clearTextByte, userinfo); err != nil {
		Logger.Error("unmarshal user info failed: ", err)
		return nil, http.StatusBadRequest, err
	}

	jsonInfo, err := json.Marshal(userinfo)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	var buf = new(bytes.Buffer)
	buf.Write(jsonInfo)

	return buf, http.StatusOK, nil
}

func GetFromNode2(req []byte) (jsonRes *bytes.Buffer, s int, err error) {
	info := new(ResponseInfo)
	err = json.Unmarshal(req, info)
	if err != nil {
		Logger.Debug("Unmarshal Fail: ", err)
		return nil, http.StatusBadRequest, err
	}

	dataChan := make(chan *[]shamir.Secret, 5)
	defer close(dataChan)
	if len(info.NodeNumber) < 3 {
		Logger.Error("needkey < 3")
		return nil, http.StatusBadRequest, errors.New("number len error")
	}
	isclose := false

	for _, v := range info.NodeNumber {
		go func(number uint32) {
			var key uint32 = number
			node := globConsistent.GetFromKey(key)
			if node == nil {
				Logger.Error("get node failed")
				dataChan <- nil
				return
			}

			var reply = new([]shamir.Secret)

			var bufKey bytes.Buffer
			bufKey.Write(info.Hash)
			bufKey.Write([]byte(":"))
			bufKey.Write(info.Addr)
			bufKey.Write([]byte(":"))
			bufKey.WriteString(strconv.FormatUint(uint64(key), 10))

			h := sha1.New()
			io.WriteString(h, bufKey.String())

			err = node.rpcNode.cli.Call("RpcNode.Get", h.Sum(nil), reply)
			if err != nil {
				Logger.Error("Get Failed:", err)
				if err == rpc.ErrShutdown {
					node.closed = true
					node.conn.Close()
				}

				dataChan <- nil
				return
			}
			if isclose {
				return
			}

			dataChan <- reply
		}(v)
	}

	var secrets [][]shamir.Secret
	for v := range dataChan {
		if v == nil {
			return
		}

		if len(secrets) >= 3 {
			isclose = true
			break
		}

		secrets = append(secrets, *v)
	}

	clearTextInt := shamir.RecoverShamirSecretsMul(secrets)
	var clearTextByte []byte
	for _, a := range clearTextInt {
		clearTextByte = append(clearTextByte, a.Bytes()...)
	}

	var userinfo = new(userInfo)
	if err = json.Unmarshal(clearTextByte, userinfo); err != nil {
		Logger.Error("unmarshal user info failed: ", err)
		return nil, http.StatusBadRequest, err
	}

	jsonInfo, err := json.Marshal(userinfo)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	var buf = new(bytes.Buffer)
	buf.Write(jsonInfo)

	return buf, http.StatusOK, nil

}
