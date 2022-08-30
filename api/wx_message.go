package api

import (
	"codeCount_exporter/base"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CodeCount struct {
	User string
	Add  int
	Del  int
}

func Query(queryStr string) (data model.Value, err error) {
	client, err := api.NewClient(api.Config{
		Address: base.Conf.App.PrometheusAddr,
	})
	if err != nil {
		base.Log.Errorf("Error creating client: %v\n", err)
		fmt.Printf("Error creating client: %v\n", err)
		return nil, err
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.Query(ctx, queryStr, time.Now())
	if err != nil {
		base.Log.Errorf("Error querying Prometheus: %v\n", err)
		fmt.Printf("Error querying Prometheus: %v\n", err)
		return nil, err
	}
	if len(warnings) > 0 {
		base.Log.Warnf("Warnings: %v\n", err)
		fmt.Printf("Warnings: %v\n", warnings)
	}

	fmt.Printf("Result:\n%v\n", result)
	return result, nil
}

func GetPromeQueryData(data string, addOrDel string) (d []CodeCount) {
	lineSlice := strings.Split(data, "\n")
	re := regexp.MustCompile("{user=\"(?P<User>[^\"]+)\"} => (?P<Count>[0-9]+) @.*")
	for _, line := range lineSlice {
		match := re.FindStringSubmatch(line)
		//groupNames := re.SubexpNames()
		if len(match) != 3 {
			continue
		}
		count, _ := strconv.Atoi(match[2])
		if addOrDel == "add" {
			d = append(d, CodeCount{
				User: match[1],
				Add:  count,
			})
		} else if addOrDel == "del" {
			d = append(d, CodeCount{
				User: match[1],
				Del:  count,
			})
		}

	}
	return d
}

func Daily(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	currentTime := fmt.Sprintf("%dh%dm", t.Hour(), t.Minute())
	queryYersAddTotal := fmt.Sprintf(`sort_desc(sum(max_over_time(add_code_count_total{branch="all"}[1d] offset %s) - min_over_time(add_code_count_total{branch="all"}[1d] offset %s) > 0) by (user))`, currentTime, currentTime)
	fmt.Printf(queryYersAddTotal)
	addData, err := Query(queryYersAddTotal)
	if err != nil {
		fmt.Fprintln(w, "api调用失败")
		return
	}
	addCodeData := GetPromeQueryData(addData.String(), "add")
	t = time.Now()
	currentTime = fmt.Sprintf("%dh%dm", t.Hour(), t.Minute())
	queryYersDelTotal := fmt.Sprintf(`sort_desc(sum(max_over_time(del_code_count_total{branch="all"}[1d] offset %s) - min_over_time(del_code_count_total{branch="all"}[1d] offset %s) > 0) by (user))`, currentTime, currentTime)
	delData, err := Query(queryYersDelTotal)
	if err != nil {
		fmt.Fprintln(w, "api调用失败")
		return
	}
	delCodeData := GetPromeQueryData(delData.String(), "del")

	for _, delCode := range delCodeData {
		// 判断在不在代码新增列表中
		if status, index := InCodeCountSlice(delCode.User, addCodeData); status {
			addCodeData[index].Del = delCode.Del
		} else {
			addCodeData = append(addCodeData, delCode)
		}
	}

	err = SendWxMsg(addCodeData)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}

}

func InCodeCountSlice(user string, codeCount []CodeCount) (exists bool, index int) {
	for i, d := range codeCount {
		if d.User == user {
			return true, i
		}
	}
	return false, -1
}

type WxCardMessage struct {
	Msgtype      string       `json:"msgtype"`
	TemplateCard TemplateCard `json:"template_card"`
}
type Source struct {
	IconURL   string `json:"icon_url"`
	Desc      string `json:"desc"`
	DescColor int    `json:"desc_color"`
}
type MainTitle struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}
type EmphasisContent struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}
type JumpList struct {
	Type  int    `json:"type"`
	URL   string `json:"url"`
	Title string `json:"title"`
}
type CardAction struct {
	Type     int    `json:"type"`
	URL      string `json:"url"`
	Appid    string `json:"appid"`
	Pagepath string `json:"pagepath"`
}
type TemplateCard struct {
	CardType              string          `json:"card_type"`
	Source                Source          `json:"source"`
	MainTitle             MainTitle       `json:"main_title"`
	EmphasisContent       EmphasisContent `json:"emphasis_content"`
	HorizontalContentList []Content       `json:"horizontal_content_list"`
	JumpList              []JumpList      `json:"jump_list"`
	CardAction            CardAction      `json:"card_action"`
}

type Content struct {
	Keyname string `json:"keyname"`
	Value   string `json:"value"`
}

func SendWxMsg(c []CodeCount) (err error) {
	var contentList []Content
	addCount := 0
	for _, count := range c {
		s := Content{
			Keyname: count.User,
			Value:   fmt.Sprintf("+%d -%d [总计%d]", count.Add, count.Del, count.Add-count.Del),
		}
		contentList = append(contentList, s)
		addCount += count.Add
	}

	yersterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	repoNum := len(base.Conf.App.GitRepo)

	var HorizontalContentList []Content
	if len(contentList) > 6 {
		HorizontalContentList = contentList[:6]
	} else if len(contentList) == 0 {
		contentList = append(contentList, Content{
			Keyname: "null",
			Value:   "null",
		})
	} else {
		HorizontalContentList = contentList
	}

	data := WxCardMessage{
		Msgtype: "template_card",
		TemplateCard: TemplateCard{
			CardType: "text_notice",
			Source: Source{
				IconURL:   "https://download.kkwen.cn/img/background/%E9%BE%99%E7%8C%AB%E5%B0%8F.png",
				Desc:      "代码统计报告",
				DescColor: 0,
			},
			MainTitle: MainTitle{
				Title: fmt.Sprintf("昨日 %s", yersterday),
				Desc:  fmt.Sprintf("统计%d个代码仓库", repoNum),
			},
			EmphasisContent: EmphasisContent{
				Title: fmt.Sprintf("%d", addCount),
				Desc:  "昨日代码新增总行数",
			},
			HorizontalContentList: HorizontalContentList,
			JumpList: []JumpList{
				{
					Type:  0,
					URL:   base.Conf.App.GrafanaUrl,
					Title: "点击卡片查看完整排行",
				},
			},
			CardAction: CardAction{
				Type:     1,
				URL:      base.Conf.App.GrafanaUrl,
				Appid:    "APPID",
				Pagepath: "PAGEPATH",
			},
		},
	}
	client := http.Client{
		Timeout: time.Second * 10,
	}
	dataBytes, _ := json.Marshal(data)

	resp, err := client.Post(base.Conf.App.WebHookUrl, "application/json", strings.NewReader(string(dataBytes)))
	if err != nil {
		base.Log.Errorf("Error call qyweixin: %v\n", err)
		return err
	}

	var wxResp = WxResp{}
	r, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(r, &wxResp)
	if wxResp.Errcode != 0 {
		base.Log.Errorf("Error call qyweixin: %v\n", wxResp.Errmsg)
		return errors.New(wxResp.Errmsg)
	}
	return nil
}

type WxResp struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}
