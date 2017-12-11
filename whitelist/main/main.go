package main

import (
	wl "github.com/nw/whitelist"
)

func main() {
	wl.CreateLogger("whitelist_httpserver.log")
	wl.InitMyDB("127.0.0.1:3306", "root", "root1234", "nw_test", "whitelist")
	wl.Start("127.0.0.1:6379")
}
