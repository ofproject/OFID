package ofworldserver

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync/atomic"
	"testing"
	//"time"
)

var (
	dd []byte
)

func getData() []byte {
	ui1 := userInfo{
		Addr: []byte("0x009cded0220f1a60a27a42212a70c3a4856783e7e839"),
		Hash: []byte("0x009cded0220f1a60a27a42212a70c3a4856783e7e838"),
		Data: []byte("0x009cded0220f1a60a27a42212a70c3a4856783e7e838"),
	}

	bq, err := json.Marshal(&ui1)
	if err != nil {
		return nil
	}

	return bq

}

/*
addr: 0x009cded0220f1a60a27a42212a70c3a4856783e7e839
Hash: 0x009cded0220f1a60a27a42212a70c3a4856783e7e838
Data: 0x009cded0220f1a60a27a42212a70c3a4856783e7e838
*/

func getData2() []byte {
	r1 := ResponseInfo{
		Addr: []byte("0x009cded0220f1a60a27a42212a70c3a4856783e7e839"),
		Hash: []byte("0x009cded0220f1a60a27a42212a70c3a4856783e7e838"),
	}

	r1.NodeNumber = append(r1.NodeNumber, 3235210122)
	r1.NodeNumber = append(r1.NodeNumber, 3235210122)
	r1.NodeNumber = append(r1.NodeNumber, 3235210122)
	r1.NodeNumber = append(r1.NodeNumber, 3235210122)
	r1.NodeNumber = append(r1.NodeNumber, 3235210122)

	br, err := json.Marshal(&r1)
	if err != nil {
		return nil
	}
	return br
}

type cstats struct {
	failedCount int64
}

func (c *cstats) addCount() {
	atomic.AddInt64(&c.failedCount, 1)
}

var (
	count = new(cstats)
)

func TestOfServer(T *testing.T) {
	body := getData()
	//request, err := http.NewRequest(http.MethodPost, "http://10.211.55.18:8082/Save", bytes.NewReader(body))
	request, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8082/Save", bytes.NewReader(body))
	if err != nil {
		T.Fatal("Request error: ", err)
		return
	}

	//request.Header.Set("Connection", "Keep-Alive")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		T.Error("handler error: ", err)
		return
	}

	//defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		T.Error("test fail")
		return
	}

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		T.Error("Get Result Failed: ", err)
		return
	}
	resp.Body.Close()

	if result != nil {
		dd = result
		//T.Log("format: ", dd)
		r := new(ResponseInfo)
		err := json.Unmarshal(result, r)
		if err != nil {
			T.Error("Unmarshal failed: ", err)
			return
		}

		/*T.Logf("result key:%s", r.Hash)
		for i, v := range r.NodeNumber {
			T.Logf("node index:%d, value:%d", i, v)
		}*/
	}

	return

}

func TestOfServerToGet(T *testing.T) {
	//body := getData2()
	//request, err := http.NewRequest(http.MethodPost, "http://10.211.55.18:8082/Get", bytes.NewReader(dd))
	request, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8082/Get", bytes.NewReader(dd))
	if err != nil {
		T.Fatal("Request error: ", err)
		return
	}
	//request.Header.Set("Connection", "Keep-Alive")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		T.Error("handler error: ", err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		T.Error("Read Err: ", err)
		return
	}

	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		T.Error("test fail")
		return
	}

	var userinfo userInfo

	err = json.Unmarshal(body, &userinfo)
	if err != nil {
		T.Error("unmarshal fail: ", err)
		return
	}

	//T.Logf("addr: %s, hash: %s, data: %s", userinfo.Addr, userinfo.Hash, userinfo.Data)

	return
}

func BenchmarkOfServer(B *testing.B) {

	body := getData()
	for i := 0; i < B.N; i++ {

		//_, err := http.Post("http://127.0.0.1:8082/Save", "application/json", buf)
		//request, err := http.NewRequest(http.MethodPost, "http://10.211.55.15:8082/Save", bytes.NewReader(body))
		//request, err := http.NewRequest(http.MethodPost, "http://10.211.55.18:8082/Save", bytes.NewReader(body))
		request, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8082/Save", bytes.NewReader(body))
		if err != nil {
			B.Fatal("Request error: ", err)
			return
		}

		request.Header.Set("Connection", "Keep-Alive")
		request.Header.Set("Content-Type", "application/json;charset=utf-8")

		var resp *http.Response
		resp, err = http.DefaultClient.Do(request)
		if err != nil {
			B.Log("handler error: ", err)
			//count.addCount()
			return
		}

		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			//count.addCount()
			return
		}
	}

	//B.Log("Failed count :", count.failedCount)
}
