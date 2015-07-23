package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	//	"io/ioutil"
	_ "github.com/djimenez/iconv-go"
	"testing"
)

func TestHttpGet(t *testing.T) {
	resp, err := HttpGet("https://www.oschina.net/home/login")
	//		resp, _ := HttpGet("http://ip138.com")
	//	resp, err := HttpGetFromProxy("http://www.ip.cn", "https://117.185.13.86:8006")
	if err != nil {
		fmt.Println(err)
	}
	//	body, _ := ioutil.ReadAll(resp.Body)
	//	utfBody, err := iconv.NewReader(resp.Body, "gbk", "utf-8")

	//	if err != nil {
	//		fmt.Println(err)
	//	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	if value, exists := doc.Find("#f_emailabc").Attr("name"); exists {
		fmt.Println(value)
	} else {
		fmt.Println("proxy unable")
	}
	//	html := doc.Text()
	//	fmt.Println(html)
	fmt.Println("编码测试")
}

func TestHttpGetSogou(t *testing.T) {
	resp, err := HttpGet("https://www.sogou.com")
	//		resp, _ := HttpGet("http://ip138.com")
	//	resp, err := HttpGetFromProxy("http://www.ip.cn", "https://117.185.13.86:8006")
	if err != nil {
		fmt.Println(err)
	}
	//	body, _ := ioutil.ReadAll(resp.Body)
	//	utfBody, err := iconv.NewReader(resp.Body, "gbk", "utf-8")

	//	if err != nil {
	//		fmt.Println(err)
	//	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	if value, exists := doc.Find("input[name=_asf]").Attr("value"); exists {
		fmt.Println(value)
	} else {
		fmt.Println("proxy unable")
	}
	//	html := doc.Text()
	//	fmt.Println(html)
	fmt.Println("编码测试")
}

func TestHttpGet360(t *testing.T) {
	resp, err := HttpGet("https://www.haosou.com")
	//		resp, _ := HttpGet("http://ip138.com")
	//	resp, err := HttpGetFromProxy("https://www.haosou.com", "https://163.125.196.79:9797")
	if err != nil {
		fmt.Println(err)
	}
	//	body, _ := ioutil.ReadAll(resp.Body)
	//	utfBody, err := iconv.NewReader(resp.Body, "gbk", "utf-8")

	//	if err != nil {
	//		fmt.Println(err)
	//	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	if value, exists := doc.Find("#search-button").Attr("value"); exists {
		fmt.Println(value)
	} else {
		fmt.Println("proxy unable")
	}
	//	html := doc.Text()
	//	fmt.Println(html)
	fmt.Println("编码测试")
}
