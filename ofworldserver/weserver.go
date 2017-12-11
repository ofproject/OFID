package ofworldserver

import (
	"bytes"
	//"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type Httserver struct {
	hs *http.Server
}

//http://127.0.0.1:8082/Save?addr=ABCDEFGHIG&hash=AADFAFDASFASD&data=AFAFDASFDASFDASFDSAFSADFSD
func save(rw http.ResponseWriter, req *http.Request) {
	//addCount()
	//time.Sleep(100 * time.Microsecond)
	if req.Method != http.MethodPost {
		rw.Write([]byte("only support post"))
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		Logger.Error("Read Err: ", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Body.Close()

	if len(body) <= 0 {
		Logger.Error("Read All Len err")
		return
	}

	var buf bytes.Buffer
	buf.Write(body)

	res, s, err := saveToNode(buf)
	if err != nil || s != http.StatusOK {
		rw.WriteHeader(400)
		return
	}

	rw.Header().Set("Content-Type", "application/json;charset=utf-8")

	if _, err := res.WriteTo(rw); err != nil {
		rw.WriteHeader(400)
		return
	}

}

func get(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		rw.Write([]byte("only support post"))
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		Logger.Error("Read Err: ", err)
		rw.WriteHeader(http.StatusBadRequest)
		return
	}

	req.Body.Close()

	if len(body) <= 0 {
		Logger.Error("Read All Len err")
		return
	}

	//var buf bytes.Buffer
	//buf.Write(body)

	res, s, err := GetFromNode2(body)
	if err != nil {
		rw.WriteHeader(s)
		return
	}

	rw.Header().Set("Content-Type", "application/json;charset=utf-8")

	if _, err := res.WriteTo(rw); err != nil {
		rw.WriteHeader(400)
		return
	}
}

func remove(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("remove"))
}

func consult(rw http.ResponseWriter, req *http.Request) {
	rw.Write([]byte("consult"))
}

func NewHttpService() *Httserver {
	http.HandleFunc("/Save", save)
	http.HandleFunc("/Get", get)
	http.HandleFunc("/Consult", consult)

	s := new(http.Server)
	s.Addr = ":8082"
	s.Handler = nil

	return &Httserver{
		hs: s,
	}
}

func (h *Httserver) Start() {
	err := h.hs.ListenAndServe()
	if err != nil {
		Logger.Error("setup httserver failed : ", err)
		os.Exit(1)
	}
}

func (h *Httserver) Close() {
	h.hs.Close()
}
