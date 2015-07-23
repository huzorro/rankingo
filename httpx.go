package main

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var headerUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/40.0.2214.111 Safari/537.36"

type Httpx struct {
	Url      string
	Headers  map[string]string
	Cookies  []*http.Cookie
	ClientIP string //本机外网IP可选
	Method   string
	ProxyUrl string //代理URL
	PostData url.Values
	Timeout  int //超时时间秒
}

func NewHttpx(reqUrl string) (h *Httpx) {
	headers := make(map[string]string)
	headers["User-Agent"] = headerUserAgent
	return &Httpx{
		Url:     reqUrl,
		Headers: headers,
		Method:  "GET",
		Timeout: 10,
	}
}

//添加header
func (h *Httpx) AddHeader(key, value string) {
	h.Headers[key] = value
}

//添加cookie
func (h *Httpx) AddCookie(c *http.Cookie) {
	h.Cookies = append(h.Cookies, c)
}

//添加POST值
func (h *Httpx) AddPostValue(key string, values []string) {
	if h.PostData == nil {
		h.PostData = make(url.Values)
	}
	if values != nil {
		for _, v := range values {
			h.PostData.Add(key, v)
		}
	}
}

//发送请求
func (h *Httpx) Send() (response *http.Response, err error) {
	if h.Url == "" {
		return nil, errors.New("URL is empty")
	}

	defer func() {
		if err != nil && h.ClientIP != "" {
			err = errors.New(err.Error() + " client ip is " + h.ClientIP)
		}
	}()

	var req *http.Request

	if h.Method == "POST" {
		req, _ = http.NewRequest("POST", h.Url, strings.NewReader(h.PostData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, _ = http.NewRequest(h.Method, h.Url, nil)
	}

	//headers
	if len(h.Headers) > 0 {
		for k, v := range h.Headers {
			req.Header.Set(k, v)
		}
	}

	//cookies
	if len(h.Cookies) > 0 {
		for _, v := range h.Cookies {
			req.AddCookie(v)
		}
	}

	transport := &http.Transport{}

	//是否使用代理
	if h.ProxyUrl != "" {
		proxy, err := url.Parse(h.ProxyUrl)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	//是否是https请求
	if h.Url[:strings.LastIndex(h.Url, "://")] == "https" {
		//跳过https证书校验
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	//设置超时时间
	dialer := net.Dialer{
		Timeout:  time.Duration(h.Timeout) * time.Second,
		Deadline: time.Now().Add(time.Duration(h.Timeout) * time.Second),
	}
	//是否使用指定的IP发送请求
	if h.ClientIP != "" {
		transport.Dial = func(network, address string) (net.Conn, error) {
			//本地地址  本地外网IP
			lAddr, err := net.ResolveTCPAddr(network, h.ClientIP+":0")
			if err != nil {
				return nil, err
			}
			dialer.LocalAddr = lAddr
			return dialer.Dial(network, address)
		}
	} else {
		transport.Dial = func(network, address string) (net.Conn, error) {
			return dialer.Dial(network, address)
		}
	}

	client := &http.Client{
		Transport: transport,
	}
	response, err = client.Do(req)
	return response, err

}

// GET请求
func HttpGet(reqUrl string) (*http.Response, error) {
	hx := NewHttpx(reqUrl)
	return hx.Send()
}

//利用指定的IP发送请求
func HttpGetFromIP(reqUrl, ipaddr string) (*http.Response, error) {
	hx := NewHttpx(reqUrl)
	hx.ClientIP = ipaddr
	return hx.Send()
}

// http GET 代理
func HttpGetFromProxy(reqUrl, proxyURL string) (*http.Response, error) {
	hx := NewHttpx(reqUrl)
	hx.ProxyUrl = proxyURL
	return hx.Send()
}

//POST请求
func HttpPost(reqUrl string, postValues map[string][]string) (*http.Response, error) {
	hx := NewHttpx(reqUrl)
	hx.Method = "POST"
	if postValues != nil {
		for k, v := range postValues {
			hx.AddPostValue(k, v)
		}
	}
	return hx.Send()
}

//POST 代理请求
func HttpPostFromProxy(reqUrl string, postValues map[string][]string, proxyURL string) (*http.Response, error) {
	hx := NewHttpx(reqUrl)
	hx.Method = "POST"
	hx.ProxyUrl = proxyURL
	if postValues != nil {
		for k, v := range postValues {
			hx.AddPostValue(k, v)
		}
	}
	return hx.Send()
}
