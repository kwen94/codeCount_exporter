package base

import (
	"bytes"
	"fmt"
	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/env"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/spf13/viper"
	"strings"
)

var (
	Conf OpsConfig
)

type OpsConfig struct {
	LogConf LogConf `json:"logConf"`
	App     App     `json:"app"`
}

type LogConf struct {
	Filename   string `json:"filename"`
	MaxSize    int    `json:"maxSize"`
	MaxBackups int    `json:"maxBackups"`
	MaxAge     int    `json:"maxAge"`
}

type App struct {
	LocalCodeDir   string `json:"localCodeDir"`
	GitRepo        []string
	GitRepoStr     string `json:"gitRepo"`
	SyncInterval   int    `json:"syncInterval"`
	PrometheusAddr string `json:"prometheusAddr"`
	GrafanaUrl     string `json:"grafanaUrl"`
	WebHookUrl     string `json:"webhookUrl"`
}

func GetConfClient() agollo.Client {
	c, err := env.InitConfig(nil)
	if err != nil {
		panic(err)
	}

	//agollo.SetLogger(&DefaultLogger{})

	client, err := agollo.StartWithConfig(
		func() (*config.AppConfig, error) {
			return c, nil
		})

	if err != nil {
		fmt.Println("err:", err)
		panic(err)
	}

	return client
}

func InitOpsConfig() OpsConfig {
	client := GetConfClient()

	content := client.GetConfig("application").GetContent()

	conf := viper.New()
	conf.SetConfigType("properties")
	err := conf.ReadConfig(bytes.NewBufferString(content))
	if err != nil {
		panic(fmt.Sprintf("viper加载配置文件错误: %s", err.Error()))
	}

	err = conf.Unmarshal(&Conf)
	if err != nil {
		panic(fmt.Sprintf("\"配置解析错误: %s", err))
	}

	if Conf.App.LocalCodeDir == "" {
		panic("目录配置获取失败")
	} else {

		err := Makedir(Conf.App.LocalCodeDir)
		if err != nil {
			panic(err.Error())
		}
	}

	if Conf.App.GitRepoStr == "" {
		panic("git配置获取失败")
	} else {
		Conf.App.GitRepo = strings.Split(Conf.App.GitRepoStr, ",")
	}

	return Conf
}
