package main

import (
//	"flag"
)

type Cfg struct {
	TaskUri       string `json:"taskUri"`
	SubmitUri     string `json:"submitUri"`
	RankPath      string `json:"rankPath"`
	RankParam     string `json:"rankParam"`
	ThreadN       int64  `json:"threadN"`
	Timeout       int64  `json:"timeout"`
	Authorization string `json:"Authorization"`
	ContentType   string `json:"Content-Type"`
	Accept        string `json:"Accept"`
}

type TaskMsg struct {
	Keyword     string `json:"keyword"`
	Destlink    string `json:"destlink"`
	Order       string `json:"order"`
	Index       string `json:"index"`
	InitTime    string `json:"initTime"`
	ExecTime    string `json:"execTime"`
	CostTime    string `json:"costTime"`
	Status      string `json:"status"`
	Description string `json:"description"`
}

func main() {
	//	rpath := flag.String("path", "rankjs", "rankjs path")
	//	threadN := flag.Int64("threadN", 5, "thread number")

}
