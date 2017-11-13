package utils

import "github.com/ethereum/go-ethereum/common"

type Application struct {
	TimeStamp  string         `json:"requestTime"`
	AAddress   common.Address `json:"from"`
	InfoRange  []string       `json:"range"`
	Share      bool           `json:"share"`
	UserTime   string         `json:"useTime"`
	ExpireTime string         `json:"expireTime"`
}

type Permission struct {
	ClearText []byte
	R         []byte
	S         []byte
}

type UserInfo struct {
	Phone   string  `json:"phone"`
	Id      string  `json:"id"`
	Address string  `json:"address"`
	Name    string  `json:"name"`

}