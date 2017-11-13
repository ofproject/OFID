package db

import (
	"testing"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
)

func TestTable_PutWhiteList(t *testing.T) {
	table, err := NewTable("")
	if err != nil {
		fmt.Println(err)

	}
	table.PutWhiteList(common.BytesToAddress([]byte{11,12,14,15,16,17,16}))
}

func TestTable_CheckWhiteList(t *testing.T) {
	table, err := NewTable("")
	if err != nil {
		fmt.Println(err)

	}
	result,err:=table.CheckWhiteList(common.BytesToAddress([]byte{11,12,14,15,16,16}))
	if err!=nil{
		fmt.Println(err)
	}
	fmt.Println(result)
}