package db


import (
	"sync"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"fmt"
	"encoding/json"
	"OfworldServer/utils"
	"github.com/ethereum/go-ethereum/rlp"

	"bytes"
)

const (
	version =1
	alive =1
	notalive =0
)
type Table struct {
	mutex sync.Mutex
    db *nodedb
}



func NewTable(nodbpath string) (*Table,error){
    db,err :=newNodedb(nodbpath,version)
	if err!=nil {
		return nil,err
	}
	tab:=&Table{
		db:db,
	}
	return tab,nil
}

func(table *Table)RegisterUserInfo(address common.Address,userInfo []byte) error{
	table.mutex.Lock()
	defer func() {
		table.mutex.Unlock()
		table.db.close()
	}()
	if err:=table.db.putUserInfo(address[:],userInfo);err!=nil{
		log.Error("Failed to save user information:%v",err)
		return err
	}

	log.Info("save user information successfully:%v",address)
	return nil
}

func(table *Table)GetUserInfo(address common.Address) ([]byte,error){
	table.mutex.Lock()
	defer func() {
		table.mutex.Unlock()
		table.db.close()
	}()
	if userInfo,err:=table.db.getUserInfo(address[:]);err!=nil{
		log.Error("Failed to get user information:%v",err)
		return nil,err
	}else{
		log.Info("get user information successfully:%v",address)
		return userInfo,nil
	}


}
//Todo:获取申请列表
func (table *Table)GetApplications(bAddress common.Address)([]byte,error) {
	table.mutex.Lock()
	defer func() {
		table.mutex.Unlock()
		table.db.close()
	}()
	applications,err:=table.db.getAppliactions(bAddress[:])
	if err!=nil{
		return nil,err
	}
	return  applications,nil
}

//Todo:存申请
func (table *Table) PutAppliation(bAddress common.Address, application utils.Application) error {

	table.mutex.Lock()
	defer func() {
		table.mutex.Unlock()
		table.db.close()
	}()

	//获取之前所有的请求，如果有就往后加没有申请新的array
	var applicationArray []utils.Application
	exsitApplications, getErr := table.db.getAppliactions(bAddress[:])
	if getErr != nil {
		if getErr.Error() == "leveldb: not found" {
			applicationArray = make([]utils.Application, 1)
			applicationArray[0] = application
		}
	} else {
		if toArrayErr := json.Unmarshal(exsitApplications, &applicationArray); toArrayErr != nil {
			fmt.Println(toArrayErr)
		}
		applicationArray = append(applicationArray, application)
	}

	textJson, jsonErr := json.Marshal(applicationArray)
	if jsonErr != nil {
		fmt.Println(jsonErr)
		return jsonErr
	}

	if err := table.db.putApplication(bAddress[:], textJson); err != nil {
		log.Error("Failed to save user information:%v", err)
		return err
	}

	log.Info("save user information successfully:%v", bAddress)
	return nil

}

//Todo:存请求许可
func (table *Table) PutPermission(aAddress common.Address, permission utils.Permission, index int) error {
	table.mutex.Lock()
	defer func() {
		table.mutex.Unlock()
		table.db.close()
	}()

	/**
	     解析铭文:
         bAddress,                    =>许可来源地址
		 permissionTo,                =>许可目标地址
         singleApplication.TimeStamp, =>许可请求时间
		 singleApplication.UserTime,  =>许可用时
		 rangeInfo,                   =>请求范围
		 pubKey,                      =>来源地址公钥
	 */
	var data []interface{}
	if err := rlp.DecodeBytes(permission.ClearText, &data); err != nil {
		fmt.Println(err)
		return err
	}
	//取出来源地址所有申请,找到对应的请求进行删除
	var applications []utils.Application
	exApplications, decodeErr := table.db.getAppliactions(data[0].([]byte))
	fmt.Println("ad: ", common.BytesToAddress(data[0].([]byte)))
	if decodeErr != nil {
		fmt.Println(decodeErr)
		return decodeErr
	}

	if err := json.Unmarshal(exApplications, &applications); err != nil {
		fmt.Println(err)
		return err
	}

	applications = append(applications[:index], applications[index+1:]...)
	newApplications, encodeErr := json.Marshal(applications)
	if encodeErr != nil {
		fmt.Println(encodeErr)
		return encodeErr
	}
	fmt.Println("put applications is ok!")

	if err := table.db.putApplication(data[0].([]byte), newApplications); err != nil {
		fmt.Println("put er: ", err)
	}

	//许可入库
	var permissionArray []utils.Permission
	exsitApplications, getErr := table.db.getPermission(aAddress[:])
	if getErr != nil {
		if getErr.Error() == "leveldb: not found" {
			permissionArray = make([]utils.Permission, 1)
			permissionArray[0] = permission
		}
	} else {
		if toArrayErr := json.Unmarshal(exsitApplications, &permissionArray); toArrayErr != nil {
			fmt.Println(toArrayErr)
		}
		permissionArray = append(permissionArray, permission)
	}
	permissionByte, err := json.Marshal(permissionArray)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if err := table.db.putPermission(aAddress[:], permissionByte); err != nil {
		fmt.Println("save permission failed")
		log.Error("Failed to save user permission:%v", err)
		return err
	}
	fmt.Println("save permission sucessfully")
	log.Info("save user permission successfully:%v", aAddress)
	return nil
}

func(table *Table)GetPermission(aAddress common.Address)([]byte,error){
	fmt.Println("get permission add: ",aAddress)

	table.mutex.Lock()
	defer func() {
		table.mutex.Unlock()
		table.db.close()
	}()
	applications,err:=table.db.getPermission(aAddress[:])
	if err!=nil{
		return nil,err
	}
	return  applications,nil
}

//PutWhiteList: 把在我们服务端人工手动注册过的账户记录到白名单中
func (table *Table) PutWhiteList(aAddress common.Address) error {
	table.mutex.Lock()
	defer func() {
		table.mutex.Unlock()
		table.db.close()
	}()
	var existAddress []common.Address

	whiteList, err := table.db.getWhiteList()
	if err != nil {
		if err.Error() == "leveldb: not found" {
			existAddress = make([]common.Address, 1)
			existAddress[0] = aAddress
		}
	} else {
		if jsonErr := json.Unmarshal(whiteList, &existAddress); jsonErr != nil {
			log.Error("put white list json unmarshal err: ", jsonErr)
			return err
		}
		existAddress = append(existAddress, aAddress)
	}
	fmt.Println(existAddress)

	whiteListByte, jsonMarErr := json.Marshal(existAddress)
	if jsonMarErr != nil {
		log.Error("put white list json marshal err: ", jsonMarErr)

		return jsonMarErr
	}
	if putErr := table.db.putWhiteList(whiteListByte); putErr != nil {
		log.Error("put white list db put err", putErr)
		return putErr
	}

	return nil
}

//CheckWhiteList: 在白名单中是否存在给定的地址
func (table *Table) CheckWhiteList(address common.Address) (bool, error) {
	table.mutex.Lock()
	defer func() {
		table.mutex.Unlock()
		table.db.close()
	}()
	existWhiteListByte, getErr := table.db.getWhiteList()
	if getErr != nil {
		return false, getErr
	}
	var exsitWhiteList []common.Address
	if jsonErr := json.Unmarshal(existWhiteListByte, &exsitWhiteList); jsonErr != nil {
		log.Error("get white list json unmarshal err: ", jsonErr)

		return false, jsonErr
	}
	if find := findAddress(address, exsitWhiteList); find {
		return true, nil
	}
	return false, nil

}

func (table *Table) SaveShamir(address common.Address, ips []byte) error{

	table.mutex.Lock()
	defer func() {
		table.mutex.Unlock()
		table.db.close()
	}()

	if err:=table.db.puteShamir(address[:],ips);err!=nil{
		return err
	}

	return  nil

}

func(table *Table) GetShamirIps(address common.Address) ([]byte,error){
	table.mutex.Lock()
	defer func() {
		table.mutex.Unlock()
		table.db.close()
	}()
    ipBytes,err:=table.db.GetShamirIps(address[:])

	if err!=nil{
		return nil,err
	}
	return ipBytes,nil
}


func findAddress(address common.Address,addressArray []common.Address) bool {
   fmt.Println("add: ",address,"list: ",addressArray)
	for i:=0;i<len(addressArray);i++{
	    if bytes.Equal(address[:],addressArray[i][:]){
			return true
		}
	}
	return false
}