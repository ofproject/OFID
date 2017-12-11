package shamir

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"
)

func TestGenerateShamie(t *testing.T) {

	result, err := GenerateShamire(9, 9, []byte("Hello world"))
	fmt.Println((new(big.Int).SetBytes([]byte("Hello world"))))
	if err != nil {
		fmt.Println(err)
	}

	RecoverSecret(result)
}

func Btest() {
	val := "如果你的go程序是用http包启动的web服务器，你想查看自己的web服务器的状态。这个时候就可以选择net/http/pprof。你只需要引入包_”net/http/pprof”"
	buf := bytes.NewBufferString(val)
	//result, err := GenerateShamire(9, 9, buf.Bytes())
	result, err := GenerateShamire(3, 5, buf.Bytes())
	//fmt.Println((new(big.Int).SetBytes([]byte("Hello world"))))
	if err != nil {
		fmt.Println(err)
	}

	RecoverSecret(result)
}

func BenchmarkShamie(B *testing.B) {
	val := "如果你的go程序是用http包启动的web服务器，你想查看自己的web服务器的状态。这个时候就可以选择net/http/pprof。你只需要引入包_”net/http/pprof”"
	B.Log("file len: ", len(val))
	for i := 0; i < B.N; i++ {
		Btest()
	}
}
