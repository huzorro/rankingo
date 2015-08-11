package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"html/template"
	"io/ioutil"
	"math"
	"math/rand"
	"os/exec"
	"sort"
	"strings"
	"time"
)

var funcMaps = template.FuncMap{
	"html": func(text string) template.HTML {
		return template.HTML(text)
	},
	"loadtimes": func(startTime time.Time) string {
		// 加载时间
		return fmt.Sprintf("%dms", time.Now().Sub(startTime)/1000000)
	},

	"url": func(url string) string {
		// 没有http://或https://开头的增加http://
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			return url
		}

		return "http://" + url
	},
	"add": func(a, b int64) int64 {
		// 加法运算
		return a + b
	},
	"div": func(a, b int64) float64 {
		//除法运算
		return float64(a) / float64(b)
	},
	"formatdate": func(t time.Time) string {
		// 格式化日期
		return t.Format(time.RFC822)
	},
	"nl2br": func(text string) template.HTML {
		return template.HTML(strings.Replace(text, "\n", "<br>", -1))
	},
}

func normByTime(start, end, step float64, sample int) map[int64]float64 {
	var norms []float64
	for i := 0; i < sample; i++ {
		q := rand.NormFloat64()
		if q >= start && q <= end {
			norms = append(norms, q)
		}
	}
	sort.Float64s(norms)
	timesMap := make(map[int64]float64)
	for i := start; i < end; i = i + step {
		qf := i*10 + 12
		qf = math.Floor(qf)
		var ss []float64
		for _, j := range norms {
			if j >= i && j <= i+step {
				ss = append(ss, j)
			}
		}
		timesMap[int64(qf)] = float64(len(ss)) / float64(len(norms))
	}
	return timesMap
}
func ProxyCheck(vurl, ip, port, vflag, vattr string) (bool, error) {
	resp, err := HttpGetFromProxy(vurl, fmt.Sprintf("https://%s:%s", ip, port))
	if err != nil {
		return false, err
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	if _, exists := doc.Find(vflag).Attr(vattr); !exists {
		return false, errors.New("Not foud")
	}
	return true, nil
}

func ExeCmd(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}
	bytesErr, err := ioutil.ReadAll(errPipe)
	if err != nil {
		return "", err
	}

	if len(bytesErr) != 0 {
		return "", errors.New(string(bytesErr))
	}

	bytesResult, err := ioutil.ReadAll(outPipe)
	if err != nil {
		return "", err
	}
	return string(bytesResult), nil
}

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func Utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}
