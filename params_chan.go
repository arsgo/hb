package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"
)

var paramsChan chan map[string]interface{}

func init() {
	paramsChan = make(chan map[string]interface{}, 100000)
}

func (c *httpClient) ResultChanHanlde(content []byte) (b bool, err error) {
	b = len(content) != 0
	if !b {
		return
	}
	defer func() {
		b = isSuccess(content)
	}()
	input := make(map[string]interface{})
	if strings.HasPrefix(string(content), "<?") {
		err = xml.Unmarshal(content, &input)
	} else if strings.HasPrefix(string(content), "{") || strings.HasPrefix(string(content), "[") {
		err = json.Unmarshal(content, &input)
	} else {
		err = fmt.Errorf("未知的格式：%s", string(content))
	}
	if err != nil {
		err = fmt.Errorf("解决JSON失败%s,%s",err.Error(),string(content))
		return 
	}
	fields := c.data.Params["->"]
	if fields != "*" {
		if input[fields] != nil {
			data := input[fields].(map[string]interface{})
			paramsChan <- data
		} else {
			paramsChan <- nil
		}

	} else {
		paramsChan <- input
	}
	return
}

func (c *httpClient) getFromChan() (r map[string]interface{}, err error) {
	if !strings.EqualFold(c.data.Params["<-"], "*") {
		r = make(map[string]interface{})
		return
	}
	ticker := time.NewTicker(time.Second * 2)
	select {
	case r = <-paramsChan:
		{
			if r == nil {
				err = errors.New("last request error")
			}
			return
		}
	case <-ticker.C:
		{
			err = errors.New("get data from chan timeout")
			r = make(map[string]interface{})
			return
		}
	}
	return

}
