package count

import (
	"codeCount_exporter/base"
	"fmt"
	"os"
)

var CountData map[string]map[string]map[string]map[string]int

func HandleRepo(repo string) (err error) {
	// 仓库不存在，初始化仓库，否则git pull更新代码
	err = GitCloneOrUpdate(repo)
	if err != nil {
		return err
	}

	// 获取所有分支
	branchs, err := GetAllBranch()
	if err != nil {
		return err
	}

	// 添加所有分支，便于真实代码量统计
	branchs = append(branchs, "all")

	CountData[repo] = make(map[string]map[string]map[string]int)
	// 循环所有分支
	for _, branch := range branchs {
		CountData[repo][branch] = make(map[string]map[string]int)
		// 获取该分支下提交的人名
		var users []string
		if branch == "all" {
			users, err = GetCommitUserList("--all")
		} else {
			users, err = GetCommitUserList(branch)
		}

		if err != nil {
			return err
		}
		// 循环人名，统计在该分支下，这个人提交、删除的总代码量
		for _, user := range users {
			var (
				add int
				del int
			)

			if branch == "all" {
				add, del, err = GetBranchUserCommitCount("--all", user)
			} else {
				add, del, err = GetBranchUserCommitCount(branch, user)
			}

			if err != nil {
				return err
			}

			CountData[repo][branch][user] = map[string]int{
				"add": add,
				"del": del,
			}
		}
	}
	return nil

}

func StartCount() {
	CountData = make(map[string]map[string]map[string]map[string]int, 100)
	// 循环git仓库列表
	for _, repo := range base.Conf.App.GitRepo {
		if base.Conf.App.LocalCodeDir != "" {
			err := os.Chdir(base.Conf.App.LocalCodeDir)
			if err != nil {
				panic(fmt.Sprintf("base.Conf.App.LocalCodeDir配置错误, %#v", err))
			}
			_ = HandleRepo(repo)
		} else {
			panic("base.Conf.App.LocalCodeDir配置为空")
		}
	}
	base.Log.Infof("CountData: %v", CountData)
}
