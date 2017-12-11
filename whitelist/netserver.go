package whitelist

import (
	"net/http"
	"os"
)

type httserver struct {
	hs *http.Server
}

func createJob(addr, msgType string, rt RquestType) *job {
	m := msgtype{
		addr:    addr,
		msgType: msgType,
	}

	job := new(job)
	job.rt = rt
	job.msg = append(job.msg, m)
	job.r = make(chan ReplyCode)

	return job
}

//http://127.0.0.7:8081/Check?addr="123123123123"
func Check(rw http.ResponseWriter, req *http.Request) {
	Logger.Info("Check")
	val := req.FormValue("addr")
	if len(val) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
	} else {
		Logger.Info("addr: ", val)

		job := createJob(val, "", CHECKTYPE)

		rdb := NewRedisDB()
		rdb.AddJob(job)

		rc := <-job.r

		if rc == REPLYOK {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte("addr:ture"))
		} else {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte("addr:false"))
		}
	}
}

//http://127.0.0.7:8081/Check?addr="123123123123"&type="adfsa"
func Add(rw http.ResponseWriter, req *http.Request) {
	Logger.Info("Add")
	val := req.FormValue("addr")
	vt := req.FormValue("type")
	if len(val) == 0 || len(vt) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
	}

	job := createJob(val, vt, ADDTYPE)
	rdb := NewRedisDB()
	rdb.AddJob(job)

	rc := <-job.r

	if rc == REPLYOK {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("addr:ture"))
	} else {
		rw.WriteHeader(http.StatusNotFound)
		rw.Write([]byte("addr:false"))
	}

}

func Number(rw http.ResponseWriter, req *http.Request) {
	Logger.Info("Number")

	val := req.FormValue("addr")
	vt := req.FormValue("type")
	if len(val) == 0 || len(vt) == 0 {
		rw.WriteHeader(http.StatusBadRequest)
	}

	err := Modify(val, vt)
	if err != nil {
		Logger.Error(err)
	}
	rw.Write([]byte("Number"))
}

func Delete(rw http.ResponseWriter, req *http.Request) {
	Logger.Info("Delete")
	rw.Write([]byte("Delete"))
}

func NewHttpServer() *httserver {
	http.HandleFunc("/Check", Check)
	http.HandleFunc("/Add", Add)
	http.HandleFunc("/Number", Number)
	http.HandleFunc("/Delete", Delete)

	//http.ListenAndServe("127.0.0.1:8081", nil)
	s := new(http.Server)
	s.Addr = "127.0.0.1:8081"
	s.Handler = nil

	return &httserver{
		hs: s,
	}
}

func (h *httserver) Start() {
	err := h.hs.ListenAndServe()
	if err != nil {
		Logger.Error("setup httserver failed : ", err)
		os.Exit(1)
	}
}

func (h *httserver) Close() {
	h.hs.Close()
}
