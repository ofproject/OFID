package utils

import (
	"time"
	"math/rand"
)


//ByteSplit 根据长度把数组进行分段
func ByteSplit(source []byte, eachLength int) [][]byte {

	count := 0

	des := make([][]byte, len(source)/eachLength+1)

	for eachLength*count <= len(source) {
		if eachLength*(count+1) > len(source) {
			des[count] = source[count*eachLength:]
			return des
		}
		des[count] = source[count*eachLength:eachLength*(count+1)]
		count ++
	}

	return des
}



//GenerateRandomNumber 生成随机不重复
func GenerateRandomNumber(start int, end int, count int) []int {

	if end < start || (end-start) < count {
		return nil
	}

	nums := make([]int, 0)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {

		num := r.Intn((end - start)) + start

		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}

		if !exist {
			nums = append(nums, num)
		}
	}

	return nums
}