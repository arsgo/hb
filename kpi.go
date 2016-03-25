package main

import (
	"fmt"
)

type requestKPI struct {
	totalSpanTime    int
	requestCount     int
	successCount     int
	failedCount      int
	argSpendTime     float32
	successPerSecond float32
	requestPerSecond float32
}

func calculateKPI(finishResponse []*response, totalMillisecond int) {

	kpi := requestKPI{}
	urlKPI := make(map[string]*requestKPI)
	for _, r := range finishResponse {
		if _, ok := urlKPI[r.url]; !ok {
			urlKPI[r.url] = &requestKPI{}
		}
		currentURLKpi := urlKPI[r.url]
		kpi.totalSpanTime += r.useTime
		kpi.requestCount++

		currentURLKpi.totalSpanTime += r.useTime
		currentURLKpi.requestCount++
		if r.success {
			kpi.successCount++
			currentURLKpi.successCount++
		} else {
			kpi.failedCount++
			currentURLKpi.failedCount++
		}
	}
	kpi.argSpendTime = float32(kpi.totalSpanTime) / float32(kpi.requestCount)
	kpi.successPerSecond = float32(kpi.successCount) / float32(kpi.requestCount)
	kpi.requestPerSecond = float32(kpi.requestCount) * 1000 / float32(totalMillisecond)

	fmt.Println()
	fmt.Println("-------------------------------------------------------------------------")
	fmt.Println("总数\t成功\t平均耗时\t每秒请求数")
	fmt.Printf("%d\t%d\t%.2f\t\t%.2f\r\n",
		kpi.requestCount, kpi.successCount,
		kpi.argSpendTime, kpi.requestPerSecond)

	fmt.Println("-------------------------------------------------------------------------")
}
