package main

import (
//	"flag"
)

type Cfg struct {
	TaskUri   string `json:"taskUri"`
	SubmitUri string `json:"submitUri"`
	RankPath  string `json:"rankPath"`
	RankParam string `json:"rankParam"`
	ThreadN   int64  `json:"threadN"`
	Timeout   int64  `json:"timeout"`
}

func main() {
	//	rpath := flag.String("path", "rankjs", "rankjs path")
	//	threadN := flag.Int64("threadN", 5, "thread number")

}
