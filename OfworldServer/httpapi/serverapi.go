package httpapi

import (
	"fmt"
	"OfworldServer/db"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/crypto"
	"crypto/rand"
	"math/big"
	"github.com/ethereum/go-ethereum/rlp"
	"time"
	"OfworldServer/utils"
	"OfworldServer/server"
	"OfworldServer/shamir"
	"github.com/op/go-logging"
)

type OfWorldAPI struct {
	logger *logging.Logger
}

func NewOfWorldAPI(logger *logging.Logger) *OfWorldAPI {

	return &OfWorldAPI{
		logger: logger,
	}

}

func (of *OfWorldAPI) Test() string {
	fmt.Println("test")
	return "test"
}

func (of *OfWorldAPI) Register(pubAddress common.Address, phone, id, name, address string) error {

	user := &utils.UserInfo{
		Phone:   phone,
		Id:      id,
		Name:    name,
		Address: address,
	}

	//信息转为byte数组
	userInfoBytes, EncodeErr := rlp.EncodeToBytes(user)

	if EncodeErr != nil {
		of.logger.Errorf("Fail to translate user's information to bytes: ", EncodeErr)
		return EncodeErr
	}
	//由于计算精度 秘钥不可以过大，所以进行分段处理
	splitByte := utils.ByteSplit(userInfoBytes, 20)

	//获取分段的秘钥匙
	secrets, shamirErr := shamir.GenerateShamirMul(3, 3, splitByte)
	if shamirErr != nil {
		of.logger.Errorf("Fail to generate Shamir key: ", shamirErr)
		return shamirErr
	}
	recordSavedIp := make([]string, len(secrets))

	//随机服务器地址
	ipRandomIndex := utils.GenerateRandomNumber(0, 4, len(secrets))

	//ips 一个随机出来的地址，一个备用地址
	var ipRandom []string
	var ipBack []string

	for i := 0; i < len(shamir.ShamirServerIp); i++ {
	IPBREAK:
		for j := 0; j < len(ipRandomIndex); j++ {
			if i == ipRandomIndex[j] {
				fmt.Println(ipRandomIndex[j])
				ipRandom = append(ipRandom, shamir.ShamirServerIp[i])
				i++
				goto IPBREAK
			}
		}
		ipBack = append(ipBack, shamir.ShamirServerIp[i])

	}

	//记录备用地址
	backIndex := 0
	var backRecorder []string

	for i := 0; i < len(secrets); i++ {

		httpClient, err := rpc.Dial(ipRandom[i])

		//如果有err 会启用备用地址,会一直在备用里面找到一台可以用的
		if err != nil {
			of.logger.Errorf("Fail to set up http client: ", err)
			httpClient.Close()

			//表示备用地址已经全本用完而且还有报错 直接结束存储
			if backIndex == len(ipBack) {
				of.logger.Error("Fail to save Shamir: ", &backIpError{hit: "all the backip is used"})
				rollback()
				return &backIpError{hit: "all the backip is used"}
			}
			resbackIndex, backErr := of.starBackIps(ipBack, backIndex, pubAddress, secrets[i])
			if backErr != nil {
				of.logger.Error("Fail to save Shamir: ", backErr)
				rollback()
				return backErr
			}
			backRecorder = append(backRecorder, ipBack[resbackIndex-1])
			backIndex = resbackIndex
			continue
		}

		//对应shamirServe的方法: SaveShamir(address common.Address, secrets []utils.Shamir),如果存储失败会启用备用地址
		callErr := httpClient.Call(nil, "shamir_saveShamir", pubAddress, secrets[i])

		if callErr != nil {
			of.logger.Errorf("Fail to save Shamir:", callErr)
			httpClient.Close()

			//启用备用ip
			if backIndex == len(ipBack) {
				of.logger.Error("Fail to save Shamir: ", &backIpError{hit: "all the backip is used"})
				rollback()
				return &backIpError{hit: "all the backip is used"}
			}
			resbackIndex, backErr := of.starBackIps(ipBack, backIndex, pubAddress, secrets[i])

			if backErr != nil {
				fmt.Println("BACKERR")
				of.logger.Error("Fail to save Shamir: ", backErr)
				rollback()
				return backErr
			}

			backRecorder = append(backRecorder, ipBack[resbackIndex-1])
			backIndex = resbackIndex
			continue
		}

		recordSavedIp[i] = ipRandom[i]

	}

	//如果存储成功，会把ip存入本机
	recordSavedIp = append(recordSavedIp, backRecorder...)

	//把存的ip入数据库
	table, tableErr := db.NewTable("")

	if tableErr != nil {
		of.logger.Errorf("Fail to open DB:", tableErr)
		return tableErr
	}

	ipbytes, EncodeErr := rlp.EncodeToBytes(recordSavedIp)

	if EncodeErr != nil {
		of.logger.Errorf("Fail to encode ipstring to bytes: ", EncodeErr)
		return EncodeErr

	}

	if tableErr := table.SaveShamir(pubAddress, ipbytes); tableErr != nil {
		of.logger.Errorf("Fail to save shamir to ip: ", tableErr)
		return tableErr
	}

	of.logger.Notice("Save Shamir Successfully: ", pubAddress.String(), recordSavedIp)

	return nil

}


//starBackIps 启动备用地址
func (of *OfWorldAPI) starBackIps(ipBack []string, inbackIndex int, pubAddress common.Address, secrets []shamir.Secret) (int, error) {
	for back := inbackIndex; back < len(ipBack); back++ {
		fmt.Println("back: ", back)
		httpClient, err := rpc.Dial(ipBack[back])
		if err != nil {
			of.logger.Errorf("Fail to set up back http client: ", inbackIndex, err)
			httpClient.Close()
			if back == len(ipBack)-1 {
				return back, &backIpError{hit: "All the backEnd is used"}
			}
			inbackIndex++
			continue
		}

		callErr := httpClient.Call(nil, "shamir_saveShamir", pubAddress, secrets)
		if callErr != nil {
			of.logger.Errorf("Fail to save Shamir back: ", inbackIndex, callErr)
			httpClient.Close()
			if back == len(ipBack)-1 {
				return back, &backIpError{hit: "All the backEnd is used"}
			}
			inbackIndex++
			continue
		}

		of.logger.Notice("take the backIP: ", inbackIndex, ":", inbackIndex)
		httpClient.Close()
		inbackIndex++
		break
	}

	return inbackIndex, nil
}


func rollback() {
	fmt.Println("I am Roll Back-------")
}
func (of *OfWorldAPI) GetShamirFromServer(address common.Address) ([][]shamir.Secret) {

	table, tableErr := db.NewTable("")

	if tableErr != nil {
		of.logger.Errorf("Fail to open DB: ", tableErr)
		return nil
	}

	ipBytes, getErr := table.GetShamirIps(address)
	fmt.Println(string(ipBytes))
	if getErr != nil {
		of.logger.Errorf("Fail to get ip from DB: ", getErr)
		return nil
	}

	var ips []string
	if decodeErr := rlp.DecodeBytes(ipBytes, &ips); decodeErr != nil {
		of.logger.Errorf("Fail to unmarshal ips to string  array:", decodeErr)
	}

	var secrets [][]shamir.Secret

	for _, ip := range ips {

		httpClient, httErr := rpc.Dial(ip)
		if httErr != nil {
			of.logger.Errorf("Fail to connect to server: ", httErr)
		}
		var singleSecret []shamir.Secret
		callErr := httpClient.Call(&singleSecret, "shamir_getShamirSecrets", address)

		if callErr != nil {
			of.logger.Errorf("Fail to get data from shamir server: ", callErr)
			httpClient.Close()
			return nil
		}

		httpClient.Close()

		secrets = append(secrets, singleSecret)
	}

	of.logger.Notice("ret")
	return secrets

}

func (of *OfWorldAPI) RequestInfo(aAddress common.Address, permissionIndex int) (utils.UserInfo, error) {
	var newUserInfo utils.UserInfo
	//取出许可
	var permissionArray []utils.Permission
	if exsitPermission, err := of.GetPermission(aAddress); err != nil {
		fmt.Println("permissionErr", err)
		return newUserInfo, err
	} else {
		if jsonErr := json.Unmarshal(exsitPermission, &permissionArray); jsonErr != nil {
			fmt.Println("permissionJson: ", jsonErr)
			return newUserInfo, nil
		}
	}
	permission := permissionArray[permissionIndex]
	//验证许可
	var data []interface{}

	/**
	     解析铭文:
         bAddress,                    =>许可来源地址
		 permissionTo,                =>许可目标地址
         singleApplication.TimeStamp, =>许可请求时间
		 singleApplication.UserTime,  =>许可用时
		 rangeInfo,                   =>请求范围
		 pubKey,                      =>来源地址公钥
	 */
	if err := rlp.DecodeBytes(permission.ClearText, &data); err != nil {
		fmt.Println(err)
		return newUserInfo, err
	}

	//验证公钥和来源是否匹配
	bAddress := common.BytesToAddress(data[0].([]byte))
	pubKey := crypto.ToECDSAPub(data[5].([]byte))
	if result := checkPubAndAddress(bAddress, *pubKey); !result {
		fmt.Println("pubKey is not match")
		return newUserInfo, &pubMatchError{"the pubKey is not match"}
	}

	//验证签名和铭文
	if signResult := ecdsa.Verify(pubKey, permission.ClearText, new(big.Int).SetBytes(permission.R), new(big.Int).SetBytes(permission.S)); !signResult {
		fmt.Println("sign is incorrect")
		return newUserInfo, &signError{"sign is incorrect"}
	}

	//取信息
	table, err := db.NewTable("")
	if err != nil {
		fmt.Println(err)
		return newUserInfo, err
	}
	userInfoByte, getErr := table.GetUserInfo(bAddress)
	if getErr != nil {
		fmt.Println("getErr", getErr)
		return newUserInfo, getErr
	}

	var userInfo utils.UserInfo

	jsonErr := json.Unmarshal(userInfoByte, &userInfo)
	if jsonErr != nil {
		fmt.Println("JsonErr", jsonErr)
		return newUserInfo, jsonErr
	}

	infoRange := makeJsonRange(data[4].([]interface{}))
	newUserInfo = *reSetUserInfo(infoRange, &userInfo)

	return newUserInfo, nil

}

func reSetUserInfo(infoRange []string, userInfo *utils.UserInfo) *utils.UserInfo {

	newUserInfo := &utils.UserInfo{}

	for _, info := range infoRange {
		if info == "address" {
			newUserInfo.Address = userInfo.Address
		}
		if info == "id" {
			newUserInfo.Id = userInfo.Id
		}
		if info == "name" {
			newUserInfo.Name = userInfo.Name
		}
		if info == "phone" {
			newUserInfo.Phone = userInfo.Phone
		}
	}

	return newUserInfo

}

func (of *OfWorldAPI) PutWhiteList(aAddress common.Address) error {

	table, err := db.NewTable("")

	if err != nil {
		fmt.Println(err)
		return err
	}

	if err := table.PutWhiteList(aAddress); err != nil {
		return err
	}

	return nil

}

//Todo:申请查询请求
func (of *OfWorldAPI) ApplyForThePremission(clearText, r, s []byte) error {

	/**
	    解析明文:
        new(big.Int).SetInt64(timeStamp).Bytes(), =》 时间戳
		aAddress,  =》  请求地址
		bAddress,  =》  目标地址
		infoRange, =》  请求信息范围
		share,     =》  是否公开自己信息
		useTime,   =》  请求时间
		pubKey,    =》  请求人的公钥
	 */
	var data []interface{}

	if err := rlp.DecodeBytes(clearText, &data); err != nil {
		fmt.Println(err)
		return err
	}

	//取出铭文中的公钥匙
	pubKeyFromByte := crypto.ToECDSAPub(data[6].([]byte))

	//验证签名
	signResult := ecdsa.Verify(pubKeyFromByte, clearText, new(big.Int).SetBytes(r), new(big.Int).SetBytes(s))
	if !signResult {
		fmt.Println("sign is incorrect！")
		return &signError{hit: "The sign is incorrect!"}
	}

	//取出铭文信息，生成json格式
	jsonTime := time.Unix(new(big.Int).SetBytes(data[0].([]byte)).Int64(), 0)
	jsonFrom := common.BytesToAddress(data[1].([]byte))
	jsonRange := makeJsonRange(data[3].([]interface{}))
	jsonBool := makeJsonBool(data[4].([]byte))
	jsonUseTime := string(data[5].([]byte))
	m, _ := time.ParseDuration(jsonUseTime)
	expireTime := jsonTime.Add(m)

	aApplication := &utils.Application{
		AAddress:   jsonFrom,
		InfoRange:  jsonRange,
		Share:      jsonBool,
		TimeStamp:  jsonTime.String(),
		UserTime:   jsonUseTime,
		ExpireTime: expireTime.String(),
	}

	//存信息
	table, err := db.NewTable("")
	if err != nil {
		fmt.Println(err)
		return err
	}

	if err := table.PutAppliation(common.BytesToAddress(data[2].([]byte)), *aApplication); err != nil {
		fmt.Println("put err", err)
		return err
	}
	return nil

}

func (of *OfWorldAPI) SavePermission(aAddress common.Address, clearText, r, s []byte, index int) error {

	//生成许可对象
	permission := &utils.Permission{
		ClearText: clearText,
		R:         r,
		S:         s,
	}
	table, err := db.NewTable("")
	if err != nil {
		fmt.Println(err)
		return err
	}
	//存许可
	if err := table.PutPermission(aAddress, *permission, index); err != nil {
		fmt.Println(err)
		return err
	}
	return nil

}

//todo:API:获取申请列表
func (of *OfWorldAPI) ApplicationList(bAddress common.Address) ([]byte, error) {

	table, err := db.NewTable("")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	applications, appErr := table.GetApplications(bAddress)
	if appErr != nil {
		fmt.Println(appErr)
		return nil, appErr
	}

	return applications, nil

}

func (of *OfWorldAPI) GetPermission(address common.Address) ([]byte, error) {

	table, err := db.NewTable("")
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	permission, getErr := table.GetPermission(address)
	if getErr != nil {
		fmt.Println(getErr)
		return nil, getErr
	}

	return permission, nil
}

func makeJsonRange(data []interface{}) []string {

	infoRange := make([]string, 0)

	for _, text := range data {
		infoRange = append(infoRange, string(text.([]byte)))
	}

	return infoRange
}

func makeJsonBool(data []byte) bool {

	if data == nil {
		return false
	}
	return true

}
func checkPubAndAddress(address common.Address, key ecdsa.PublicKey) bool {

	pubToAddress := crypto.PubkeyToAddress(key)

	if pubToAddress == address {
		return true
	}
	return false

}

func generateSign(context interface{}, key *ecdsa.PrivateKey) (clearText []byte, r *big.Int, s *big.Int, err error) {

	contextBytes, err := rlp.EncodeToBytes(context)
	if err != nil {
		fmt.Println(err)
		return nil, nil, nil, err
	}
	r, s, signErr := ecdsa.Sign(rand.Reader, key, contextBytes)

	if signErr != nil {

		return contextBytes, nil, nil, err

	}

	return contextBytes, r, s, nil
}
