package httpapi

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/op/go-logging"
	"ShamirServer/db"
	"ShamirServer/utils"
	"encoding/json"
)

type ShamirAPI struct {
    logger *logging.Logger
}



func NewShamirAPI(logger *logging.Logger) *ShamirAPI {

	return &ShamirAPI{
		logger:logger,
	}

}


func(api *ShamirAPI)SaveShamir(address common.Address, secrets []utils.Shamir) error{

	table,err:=db.NewTable("")

	if err!=nil{
		api.logger.Error("Fail to open db: ",err)
		return err

	}

     jsonShamir,jsonErr:=json.Marshal(secrets)

	if jsonErr!=nil{
		api.logger.Errorf("failed to marshal secrets to json")
		return jsonErr
	}

	if dbErr :=table.PutShamir(address,jsonShamir);dbErr!=nil{
		api.logger.Error("Fail to save shamir: ",err)
		return dbErr
	}

     api.logger.Notice("ShamirSaved: ",address.String(),"secrets: ",jsonShamir)

	return nil

}

func(api *ShamirAPI) GetShamirSecrets(address common.Address) ([]utils.Shamir){

	var secrets []utils.Shamir
	table,err:=db.NewTable("")

	if err!=nil{
		api.logger.Errorf("Fail to Open DB: ",err)
      return secrets
	}

	secretsByte,getErr:=table.GetShamirSecrets(address)
	if getErr!=nil{
		api.logger.Errorf("Fail to get secrets: ",secrets)
		return secrets
	}

	if jsonErr:=json.Unmarshal(secretsByte,&secrets);jsonErr!=nil{
		api.logger.Errorf("Fail to unmarshal: ",jsonErr)
		return secrets
	}

   return  secrets

}


