package db

import (
	"github.com/syndtr/goleveldb/leveldb"

	"encoding/binary"
	"bytes"
	"os"

	"fmt"
)

var (
	addressShamir =":shamir"
	addressIP =":IP"
	addressOnLine=":online"
	addressUserInfo=":userInfo"
	addresApplication=":application"
	addressPermission=":permission"
	addressWhiteList="whitelist"
	versionpre = []byte("version")
	defaultPath="/Users/tianyun/Documents/leveltestdb"
)
type nodedb struct {
	filename string
    db     *leveldb.DB

}


func newNodedb(path string, version int)(*nodedb,error){
	if path == ""{
	  path = defaultPath
	}
	db,err :=leveldb.OpenFile(path,nil)
	if err!=nil {
		return nil,err
	}
	currentVer := make([]byte, binary.MaxVarintLen64)
	//putvarint return number of writer
	currentVer = currentVer[:binary.PutVarint(currentVer, int64(version))]

	versionvalue,getError :=db.Get(versionpre,nil)
	switch getError {
	case leveldb.ErrNotFound:
		if err:=db.Put(versionpre,currentVer,nil);err!=nil{
			db.Close()
			return nil,err
		}

	case nil:
		if !bytes.Equal(versionvalue,currentVer){
			db.Close()
			if err=os.RemoveAll(path);err!=nil {
				return nil,err
			}
			return newNodedb(path,version)
		}
	}
	return &nodedb{
		db:       db,
		filename: path,
	},nil
}

func(db *nodedb)close(){
	db.db.Close()
}



func makeKey(address []byte,filed string) []byte{
	fmt.Println("make ad",address)
	fmt.Println("make filed",filed)
	return append(address,filed...)
}

func(db *nodedb)putWhiteList(address []byte) error{
	if err:=db.db.Put([]byte(addressWhiteList),address,nil);err!=nil{
		return err
	}
	return nil
}
func(db *nodedb)getWhiteList() ([]byte,error){
	 whiteList,err:=db.db.Get([]byte(addressWhiteList),nil)
	 if err!=nil{
		 return nil,err
	 }
	return whiteList,nil
}
func(db *nodedb)getAddressFromWhiteList(address []byte) error{
	if err:=db.db.Put([]byte(addressWhiteList),address,nil);err!=nil{
		return err
	}
	return nil
}
func(db *nodedb)putUserInfo(address []byte, userInfo []byte)error{
	if err:=db.db.Put(makeKey(address,addressUserInfo),userInfo,nil);err!=nil{
		return err
	}
	return nil
}

func(db *nodedb)getUserInfo(address []byte)([]byte,error){
	if userInfo,err:=db.db.Get(makeKey(address,addressUserInfo),nil);err!=nil{
		return nil,err
	}else{
       return userInfo,nil
	}
}


func(db *nodedb)putApplication(address []byte,application []byte) error{
	if err:=db.db.Put(makeKey(address,addresApplication),application,nil);err!=nil{
		return err
	}
	return nil
}

func(db *nodedb)getAppliactions(address []byte)([]byte,error){
	applications,err:=db.db.Get(makeKey(address,addresApplication),nil)
	if err!=nil{
		return nil,err
	}
	return applications,nil
}

func(db *nodedb)putPermission(address []byte,permission []byte)error{
	if err:=db.db.Put(makeKey(address,addressPermission),permission,nil);err!=nil{
		return err
	}
	return nil
}

func (db *nodedb)getPermission(address[]byte)([]byte,error){
	permission,err:=db.db.Get(makeKey(address,addressPermission),nil)
	if err!=nil{
		return nil,err
	}
	return permission,nil
}

func (db *nodedb)puteShamir(address[]byte, ips []byte) error {
	if err:=db.db.Put(makeKey(address,addressShamir),ips,nil);err!=nil{
		return err
	}
	return nil
}

func (db *nodedb)GetShamirIps(address []byte) ([]byte,error){
	ipbytes,getErr:=db.db.Get(makeKey(address,addressShamir),nil)
	if getErr!=nil{
		return nil,getErr
	}
	return ipbytes,nil
}