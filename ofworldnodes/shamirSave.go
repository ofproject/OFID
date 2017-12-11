package ofworldnodes

import (
	"encoding/json"
	"github.com/nw/ofworldnodes/db"
	"math/big"
)

type Shamir struct {
	Item   int64
	Result *big.Int
}

type ShamirAPI uint8

func (api *ShamirAPI) SaveShamir(address []byte, secrets []Shamir) error {

	table, err := db.NewTable("")

	if err != nil {
		Logger.Error("Fail to open db: ", err)
		return err

	}

	jsonShamir, jsonErr := json.Marshal(secrets)

	if jsonErr != nil {
		Logger.Errorf("failed to marshal secrets to json")
		return jsonErr
	}

	if dbErr := table.PutShamir(address, jsonShamir); dbErr != nil {
		Logger.Error("Fail to save shamir: ", err)
		return dbErr
	}

	//Logger.Notice("ShamirSaved: ", address, "secrets: ", jsonShamir)

	return nil

}

func (api *ShamirAPI) GetShamirSecrets(address []byte) []Shamir {

	var secrets []Shamir
	table, err := db.NewTable("")

	if err != nil {
		Logger.Errorf("Fail to Open DB: ", err)
		return secrets
	}

	secretsByte, getErr := table.GetShamirSecrets(address)
	if getErr != nil {
		Logger.Errorf("Fail to get secrets: ", secrets)
		return secrets
	}

	if jsonErr := json.Unmarshal(secretsByte, &secrets); jsonErr != nil {
		Logger.Errorf("Fail to unmarshal: ", jsonErr)
		return secrets
	}

	return secrets

}
