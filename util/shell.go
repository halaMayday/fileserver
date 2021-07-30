package util

import (
	"bytes"
	"os/exec"
)

//执行linux命令
func ExecLinuxShell(s string) (string, error) {
	//函数类型返回一个io.writer类型的 *Cmd
	cmd := exec.Command("/bin/bash", "-c", s)
	//通过bytes.Buffer將byte类型转为string类型
	var result bytes.Buffer
	cmd.Stdout = &result

	//Run执行cmd包含的命令，并阻塞直至完成
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return result.String(), err
}
