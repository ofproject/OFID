package httpapi

import (
	"testing"
	"fmt"
	"OfworldServer/utils"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
)

func TestOfWorldAPI_Request(t *testing.T) {
	logger:=utils.CreateLogger()

	addbig,bo:=new(big.Int).SetString("8f21e0b2098de44dfb50ad7e7e3cc8718c7343a6",16)
	if bo{

	}

	add :=common.BytesToAddress(addbig.Bytes())
	fmt.Println(add.String())
	apiInstance :=NewOfWorldAPI(logger)
	apiInstance.Register(add,"15902126200","330381199305210933","陈天赟","北京市海淀区财智大厦2楼孵化器")


}

func TestOfWorldAPI_GetShamirFromServer(t *testing.T) {
	logger:=utils.CreateLogger()

	addbig,bo:=new(big.Int).SetString("8f21e0b2098de44dfb50ad7e7e3cc8718c7343a6",16)
	if bo{

	}

	add :=common.BytesToAddress(addbig.Bytes())
	fmt.Println(add.String())

	apiInstance :=NewOfWorldAPI(logger)
    secrets:=apiInstance.GetShamirFromServer(add)
	fmt.Println(secrets)
}



