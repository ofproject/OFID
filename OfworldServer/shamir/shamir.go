package shamir

import (
	"math/rand"
	"time"
	"math/big"
	"fmt"
	"OfworldServer/utils"
)

var (
	randSeed = rand.NewSource(time.Now().Unix())
	ShamirServerIp= []string{
		"http://13.58.18.141:8080",
		"http://18.221.37.164:8080",
		"http://18.221.56.4:8080",
		"http://13.58.183.250:8080",
		"http://18.220.116.44:8080",
	}
)



type Secret struct {

	Item   int64
	Result *big.Int

}

func newSecret(item int64, result *big.Int) *Secret {

	return &Secret{
		Item:   item,
		Result: result,
	}

}
//GenerateShamirMul 沙米尔分段加密， 传入的数据为分段的bytes 二维数组=》返回的形式也是二维的秘钥
func GenerateShamirMul(keyNeed, keyNumber int, clearText [][]byte) ([][]Secret, error) {

	mulSecret := make([][]Secret, keyNumber)
	for i := 0; i < len(clearText); i++ {
		secrets, err := GenerateShamire(keyNeed, keyNumber, clearText[i])
		if err != nil {
			return nil, err
		}

		for j := 0; j < keyNumber; j++ {
			innerArray := mulSecret[j]
			innerArray = append(innerArray, secrets[j])
			mulSecret[j] = innerArray
		}
	}
      fmt.Println(mulSecret)
	return mulSecret, nil

}


//GenerateShamire 对外生成沙米尔group
func GenerateShamire(keyNeed, keyNumber int, clearText []byte) ([]Secret, error) {

	if keyNeed > keyNumber {
		return nil, &keyNumberError{hit: "The keyNeed is larger than keyNumber"}
	}

	poly := initPolynomial(keyNeed)
	secretGroup := generateSecret(poly, keyNumber, new(big.Int).SetBytes(clearText))

	return secretGroup, nil
}

//initPolynomial 初始化要用的多项式，最大次数等于keyNeed-1
func initPolynomial(keyNeed int) map[int]int64 {

	polynomial := make(map[int]int64)
	for i := 1; i <= keyNeed-1; i++ {
		item := randSeed.Int63()
		polynomial[i] = item
	}
	return polynomial

}

//generateSecret 根据需要的多项式和Number生成门钥匙
func generateSecret(poly map[int]int64, secretNumber int, clearText *big.Int) []Secret {

	secretArray := make([]Secret, secretNumber)

	randNumber :=utils.GenerateRandomNumber(1,100,secretNumber)

	for i := 0; i < secretNumber; i++ {
		x := randNumber[i]
		result := new(big.Int).SetInt64(0)
		for pow, item := range poly {
			stepResult := calculateEachItem(int64(x), item, pow)
			result.Add(result, stepResult)
		}
		result.Add(result, clearText)
		secretArray[i] = *newSecret(int64(x), result)

	}

	return secretArray

}

//calculateEachItem  计算item*x^pow
func calculateEachItem(x int64, item int64, pow int) *big.Int {

	itemPow := new(big.Int).Exp(new(big.Int).SetInt64(x), new(big.Int).SetInt64((int64(pow))), nil)
	return new(big.Int).Mul(new(big.Int).SetInt64(item), itemPow)

}

//recoverSecret 恢复门限
func RecoverSecret(secretGroup []Secret) {

	finalResult := new(big.Float).SetFloat64(0.0).SetPrec(500)
	for i := 0; i < len(secretGroup); i++ {
		stepMul := new(big.Float).SetFloat64(1.0).SetPrec(500)
		for j := 0; j < len(secretGroup); j++ {
			if j != i {
				stepMul.Mul(stepMul, interpolationDiv(secretGroup[i].Item, secretGroup[j].Item))
			}
		}
		stepMul.Mul(stepMul, new(big.Float).SetInt(secretGroup[i].Result))
		finalResult.Add(finalResult, stepMul)
		//解出来的值要么就是.999999999要么就是.000000000000，加0.1 取整
		finalResult.Add(finalResult, new(big.Float).SetFloat64(0.1))
	}

	finalInt := new(big.Int)
	finalResult.Int(finalInt)
	fmt.Println(finalResult.Int(finalInt))

}

//interpolationDiv y/(x-y)
func interpolationDiv(x, y int64) *big.Float {

	subStep := new(big.Int).Sub(new(big.Int).SetInt64(x), new(big.Int).SetInt64(y))
	return new(big.Float).Quo(new(big.Float).SetInt64(y).SetPrec(500), new(big.Float).SetInt(subStep).SetPrec(500))

}

