package main

import (
	"codeCount_exporter/api"
	"codeCount_exporter/base"
	"codeCount_exporter/count"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"os"
	"time"
)

func main() {

	conf := base.InitOpsConfig()
	fmt.Printf("%v\n", base.Conf)

	base.InitLogger(&conf.LogConf)

	// 统计初始数据
	time.Sleep(time.Second * time.Duration(3))
	go count.StartCount()
	workerGit := count.NewClusterManager()

	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(workerGit)

	gatherers := prometheus.Gatherers{
		//prometheus.DefaultGatherer, // prometheus自带的监控项
		reg,
	}

	api.InitRoute(gatherers)


	go func() {
		for {
			time.Sleep(time.Second * time.Duration(conf.App.SyncInterval))
			count.StartCount()
		}
	}()

	if err := http.ListenAndServe(":7777", nil); err != nil {
		os.Exit(1)
	}

}
