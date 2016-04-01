package main

import (
	"flag"
	"github.com/colinyl/lib4go/logger"
)

var log *logger.Logger

func init() {
	log, _ = logger.New("hb",true)
}

//批量下单，指定进程数，创建指定进程的程序，并批量进行下单请求
func main() {

	var (
		totalRequest int
		concurrent   int
		timeout      int
		configPath   string
		requestURL   string
		sleep        int
	)

	flag.IntVar(&totalRequest, "n", 0, "总请求个数")
	flag.IntVar(&concurrent, "c", 1, "并发处理数")
	flag.IntVar(&timeout, "t", 0, "超时时长，默认不限制")
	flag.StringVar(&configPath, "f", "", "参数配置文件,未指定时需使用-u指定URL")
	flag.StringVar(&requestURL, "u", "", "请求的URL,未指定时需通过-f参数指定参数文件")
	flag.IntVar(&sleep, "s", 0, "每笔请求休息毫秒数")

	flag.Parse()

	config := NewConfig(configPath, requestURL)

	s, p := NewProcesss(totalRequest, concurrent, configPath, requestURL, timeout, sleep, config.Items)
	if !s {
		return
	}

	response, totalMillisecond := p.Start()

	calculateKPI(response, totalMillisecond)

}
