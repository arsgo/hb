package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

type response struct {
	success bool
	useTime int
	url     string
	index   int
}

type client interface {
	RunNow(int) *response
	GetLen() int
}

type process struct {
	clients      client
	startChan    chan int
	finishChan   chan *response
	totalRequest int
	concurrent   int
	timeout      int
	sleep        int
	configPath   string
	requestURL   string
	dataBlocks   []*dataBlock
}

func NewProcesss(totalRequest int, concurrent int, configPath string,
	requestURL string, timeout int, sleep int, dataBlocks []*dataBlock) (bool, *process) {
	p := &process{totalRequest: totalRequest, concurrent: concurrent,
		configPath: configPath, requestURL: requestURL, timeout: timeout, sleep: sleep, dataBlocks: dataBlocks}
	p.startChan = make(chan int, concurrent)
	p.finishChan = make(chan *response, totalRequest)
	return p.init(), p
}

func (p *process) init() bool {
	if strings.EqualFold(p.configPath, "") && strings.EqualFold(p.requestURL, "") {
		flag.Usage()
		return false
	}
	//初始化消息通道，并初始化工作进程数
	if p.totalRequest > 0 {
		fmt.Printf("启动 %d 个工作进程,处理 %d个请求\n", p.concurrent, p.totalRequest)
	} else {
		fmt.Printf("启动 %d 个工作进程,无限次发送请求\n", p.concurrent)
	}

	//创建http clients
	p.clients = NewHttpClients(p.concurrent, p.dataBlocks)

	for i := 0; i < p.concurrent && i < p.totalRequest; i++ {
		go p.run(p.startChan, p.finishChan)
	}
	return true
}

func (p *process) Start() ([]*response, int) {
	var (
		finishResponse   []*response
		passTime         int
		totalMillisecond int
	)
	defer close(p.startChan)
	defer close(p.finishChan)

	flowStartTime := time.Now()
	for index := 0; index < p.concurrent; index++ {
		p.startChan <- index
	}
	timePiker := time.NewTicker(time.Second)
loop:
	for {
		select {
		case f := <-p.finishChan:
			{
				finishResponse = append(finishResponse, f)
				if p.totalRequest != 0 && len(finishResponse) >= p.totalRequest {
					break loop
				}
				if p.sleep > 0 {
					time.Sleep(time.Duration(time.Millisecond * time.Duration(p.sleep)))
				}
				if len(finishResponse)+p.concurrent-1 < p.totalRequest {
					p.startChan <- len(finishResponse) % p.clients.GetLen()
					p.run(p.startChan, p.finishChan)
				}
			}
		case <-timePiker.C:
			{
				passTime++
				if passTime%2 == 0 && len(finishResponse) > 0 {
					fmt.Printf("完成请求数:%d\r\n", len(finishResponse))
				}

				if passTime >= p.timeout && p.timeout > 0 {
					break loop
				}
			}
		}
	}
	flowEndTime := time.Now()
	totalMillisecond = subTime(flowStartTime, flowEndTime)
	return finishResponse, totalMillisecond
}

func (p *process) run(startNotify chan int, finishNotify chan *response) {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r.(error).Error())
		}
	}()
	index := <-startNotify
	finishNotify <- p.clients.RunNow(index)
}
