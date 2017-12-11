package ofworldnodes

import (
	"net/rpc"
	"testing"
)

func BenchmarkOfNode(B *testing.B) {
	client1, err := rpc.DialHTTP("tcp", "127.0.0.1:8009")
	if err != nil {
		B.Error("rpc Dial failed: ", err)
		//continue
		return
	}

	for i := 0; i < B.N; i++ {
		var reply bool = false

		err = client1.Call("Node.TestSave", []byte("fdasfdsafsadfdsafs"), &reply)
		if err != nil {
			B.Error("rcp Call failed: ", err)
			continue
		}

		client1.Close()
	}
}
