package main

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func TestHttpGet(t *testing.T) {
	resp, _ := HttpGet("https://www.oschina.net/home/login")
	//	resp, _ := HttpGetFromProxy("http://ip138.com", "http://39.189.1.145:8123")
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	fmt.Println(string(body))
	fmt.Println("编码测试")
}
