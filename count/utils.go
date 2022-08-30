package count

import (
	"codeCount_exporter/base"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var GitCloneTimeout = 20
var GitPullTimeout = 20
var GitCommCommandTimeout = 10

func Execute(command string, timeout int) (result string, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "/bin/bash", "-c", command)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func GitClone(repo string) (err error) {
	cmd := "git clone -b master " + repo
	_, err = Execute(cmd, GitCloneTimeout)
	return
}

func GitCloneOrUpdate(repo string) (err error) {
	dirNameSlice := strings.Split(repo, "/")
	if len(dirNameSlice) == 0 {
		return errors.New("git地址错误, " + repo)
	}
	re, _ := regexp.Compile(`\.git$`)
	dirName := re.ReplaceAllString(dirNameSlice[len(dirNameSlice)-1], "")

	stat, err := os.Stat(dirName)
	if err != nil {
		if os.IsNotExist(err) {
			// 路径不存在，初始化
			base.Log.Infof("仓库 %s, dirName %s 不存在，初始化中...", repo, dirName)
			err = GitClone(repo)
			if err != nil {
				base.Log.Errorf("git clone 失败， gitRepo: %s, ERROR: %s", repo, err.Error())
				return err
			}
		} else {
			base.Log.Errorf("获取目录 %s 失败: %s", dirName, err.Error())
			return err
		}
	} else {
		// 判断是否是目录
		if !stat.IsDir() {
			base.Log.Errorf("初始化目录[%s]被文件占用，无法初始化仓库[%s]", dirName, repo)
			return errors.New(fmt.Sprintf("初始化目录[%s]被文件占用，无法初始化仓库[%s]", dirName, repo))
		}
	}

	// 进入代码目录
	if err := os.Chdir(dirName); err != nil {
		base.Log.Errorf("进入目录[%s]失败, ERROR: %s", dirName, err.Error())
		return err
	}

	// 拉取最新代码
	if _, err := Execute("git pull", GitPullTimeout); err != nil {
		base.Log.Errorf("git pull 失败， repo: [%s], ERROR: %s", repo, err.Error())
		return err
	}

	return nil
}

func GetAllBranch() (branchs []string, err error) {
	// 获取分支的时候，获取以remotes/origin/开头的分支名，例如 remotes/origin/dev
	// 最后生成metrics的时候，去除掉remotes/origin/
	cmd := `git branch -a | grep "remotes" | grep -v -E '\->' | awk '{print $1}'`
	r, err := Execute(cmd, GitCommCommandTimeout)
	if err != nil {
		base.Log.Errorf("查看所有分支失败, ERROR: %s", err.Error())
		return nil, err
	}

	if r == "" {
		return nil, errors.New("未获取到git分支列表")
	}

	return ParserGitCmdResult(r), nil
}

func ParserGitCmdResult(s string) (result []string) {
	rSlice := strings.Split(s, "\n")
	for _, b := range rSlice {
		n := strings.Trim(b, " ")
		if n != "" {
			result = append(result, n)
		}

	}
	return
}

func GetCommitUserList(branch string) (users []string, err error) {
	cmd := "git log " + branch + " --format='%aN' | sort -u"
	r, err := Execute(cmd, GitCommCommandTimeout)
	if err != nil {
		base.Log.Errorf("查人提交人信息失败： %s", err.Error())
		return nil, err
	}
	if r == "" {
		return
	}

	users = ParserGitCmdResult(r)
	return
}

func GetBranchUserCommitCount(branch string, user string) (add int, del int, err error) {
	cmd := fmt.Sprintf("git log %s --author='%s' --pretty=tformat: --numstat | awk 'BEGIN{add=0;del=0}{add=add+$1;del=del+$2}END{print add, del}'", branch, user)
	r, err := Execute(cmd, GitCommCommandTimeout)
	if err != nil {
		base.Log.Errorf("查看分支[%s]提交详情失败： %s", branch, err.Error())
		return 0, 0, err
	}

	resultSlice := ParserGitCmdResult(r)
	result := strings.Split(resultSlice[0], " ")
	add, _ = strconv.Atoi(result[0])
	del, _ = strconv.Atoi(result[1])
	return
}
