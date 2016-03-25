package main

import (
	"fmt"
	"os"
)

type HttpClients struct {
	clients []*httpClient
	blocks  []*dataBlock
	count   int
}

func NewHttpClients(count int, blocks []*dataBlock) *HttpClients {
	c := &HttpClients{count: count, blocks: blocks}
	c.clients = make([]*httpClient, 0)
	for i := 0; i < c.count; i++ {
		index := i % len(c.blocks)
		c.clients = append(c.clients, NewHttpClient(c.blocks[index]))
	}
	return c
}

func (c *HttpClients) HttpGet(i int) *response {
	if i > len(c.clients)-1 {
		fmt.Printf("索引错误:%d\r\n", i)
		os.Exit(1)
	}
	client := c.clients[i]
	return client.Reqeust()
}
