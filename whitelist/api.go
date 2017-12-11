package whitelist

import (
//"net/http"
)

func Start(addr string) {
	redisAddr = addr
	Logger.Info("redis addr:", redisAddr)
	rdb := NewRedisDB()

	rdb.Run()

	hs := NewHttpServer()
	hs.Start()

	rdb.Wait()
	rdb.Close()
	hs.Close()
}
