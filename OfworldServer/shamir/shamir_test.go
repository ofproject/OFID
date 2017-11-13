package shamir

import (
	"testing"
	"fmt"
	"math/big"
)

func TestGenerateShamie(t *testing.T) {

	result,err:=GenerateShamire(9,9,[]byte("Hello world"))
	fmt.Println((new(big.Int).SetBytes([]byte("Hello world"))))
	if err!=nil {
		fmt.Println(err)
	}

	RecoverSecret(result)
}
