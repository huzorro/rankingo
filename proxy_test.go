package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestKuaidailiAll(t *testing.T) {
	var (
		proxy     ProxyMsg
		kuaidaili ProxyKuaidaili
	)

	api, err := url.Parse("http://www.kuaidaili.com/api/getproxy")
	if err != nil {
		fmt.Printf("url parse fails %s", err)
		return
	}

	//	http://www.kuaidaili.com/api/getproxy/
	//	?orderid=953736356849843&num=1&area=%E4%B8%AD%E5%9B%BD
	//	&browser=1&protocol=2&method=1&an_ha=1&quality=1&sort=0&format=json&sep=1
	q := api.Query()
	//订单号
	q.Add("orderid", "953736356849843")
	//数量
	q.Add("num", "1")
	//稳定性
	q.Add("quality", "1")
	//地区
	q.Add("area", "中国")
	//协议 2:https
	q.Add("protocol", "2")
	//请求方式 get
	q.Add("method", "1")
	//游览器ua支持 pc
	q.Add("browser", "1")
	//高匿名
	q.Add("an_ha", "1")
	//结果返回格式
	q.Add("format", "json")
	//过滤当天提取过的ip
	q.Add("dedup", "1")
	api.RawQuery = q.Encode()
	for {
		for {
			resp, err := HttpGet(api.String())
			defer resp.Body.Close()
			if err != nil {
				fmt.Printf("proxy kuaidaili get fails %s", err)
				return
			}
			body, err := ioutil.ReadAll(resp.Body)
			err = json.Unmarshal(body, &kuaidaili)

			if err != nil {
				fmt.Printf("Unmarshal fails %s", err)
				return
			}
			if kuaidaili.Code == int64(0) {
				break
			}
			time.Sleep(3000 * time.Millisecond)
		}

		proxy.Ip = strings.Split(kuaidaili.Data.ProxyList[0], ":")[0]
		proxy.Port = strings.Split(kuaidaili.Data.ProxyList[0], ":")[1]
		fmt.Printf("%+v", proxy)
		resp, err := HttpGetFromProxy("https://www.sogou.com", "https://"+proxy.Ip+":"+proxy.Port)
		if err != nil {
			fmt.Printf("proxy check fails %s", err)
			continue
		}
		doc, err := goquery.NewDocumentFromResponse(resp)
		if err != nil {
			fmt.Printf("go query create document fails %s", err)
			continue
		}
		defer resp.Body.Close()
		if _, exists := doc.Find("input[name=_asf]").Attr("value"); !exists {
			fmt.Printf("can not get the specified element validation fails")
			continue
		}
		fmt.Println("validation success")
	}
	//proxy ua

	//	ua, err := HttpGetFromProxy("http://localhost:10086/api/proxy/ua", "http://"+proxy.Ip+":"+proxy.Port)
	//	if err != nil {
	//		fmt.Printf("proxy ua check fails %s", err)
	//		return
	//	}
	//	defer ua.Body.Close()

	//	uaResponse, err := ioutil.ReadAll(ua.Body)
	//	if err != nil {
	//		fmt.Printf("ua response read fails %s", err)
	//		return
	//	}

	//	fmt.Println(string(uaResponse))
}
