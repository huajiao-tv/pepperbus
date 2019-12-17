package main

import (
	"bytes"
	"io"
	"os/exec"
	"strings"
)

// path		为执行远程命令的脚本路径
// ip		远程ip
// user		远程登录用户
// pass		远程登录密码
// cmdStr	远程执行的命令
// cmdArr	本地执行命令数组，二维数组，每一个子项为一条命令，命令之间为管道符
func execCmd(local bool, path, ip, user, pass, cmdStr string, cmdArr [][]string) (string, error) {
	if local {
		return multiCmdExec(cmdArr)
	} else {
		cmd := exec.Command(path, ip, user, pass, cmdStr)
		if buf, err := cmd.Output(); err != nil {
			return "", err
		} else {
			res := strings.Split(string(buf), "\r\n")
			return res[len(res)-2], err
		}
	}
}

//
func genCmd(cmd []string) *exec.Cmd {
	if len(cmd) >= 2 {
		// 带参数命令
		return exec.Command(cmd[0], cmd[1:]...)
	} else {
		// 无参数命令
		return exec.Command(cmd[0])
	}
}

// cmdArr	本地执行命令数组，二维数组，每一个子项为一条命令，命令之间为管道符
func multiCmdExec(cmdArr [][]string) (string, error) {
	// 只有一个命令
	if len(cmdArr) == 1 {
		cmd := genCmd(cmdArr[0])
		if buf, err := cmd.Output(); err != nil {
			return string(buf), err
		} else {
			return "", err
		}
	}

	// 构造流串
	cmdCount := len(cmdArr)
	r := make([]*io.PipeReader, 0, cmdCount)
	w := make([]*io.PipeWriter, 0, cmdCount)
	for i := 0; i < cmdCount; i++ {
		rNew, wNew := io.Pipe()
		r = append(r, rNew)
		w = append(w, wNew)
	}

	// 构造流串行的命令
	cmdS := make([]*exec.Cmd, 0, cmdCount)
	// 第一个命令
	cmdOld := genCmd(cmdArr[0])
	cmdOld.Stdout = w[0]
	cmdS = append(cmdS, cmdOld)

	// 后续命令
	var res bytes.Buffer
	for idx, cmdArg := range cmdArr[1:] {
		cmdNew := genCmd(cmdArg)
		cmdS = append(cmdS, cmdNew)
		cmdNew.Stdin = r[idx]

		if idx == len(cmdArr[1:])-1 {
			cmdNew.Stdout = &res
		} else {
			cmdNew.Stdout = w[idx+1]
		}
	}

	// 执行流串行的命令
	for _, cmd := range cmdS {
		cmd.Start()
	}
	for i := 0; i < cmdCount; i++ {
		cmdS[i].Wait()
		w[i].Close()
	}

	arr := strings.Split(strings.TrimSpace(res.String()), "\n")
	return arr[0], nil
}
