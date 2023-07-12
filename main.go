package sshclient

import (
	"fmt"
	"time"
)

func main() {
	cli, err := NewClient("xx.xx.xx.xx", "22", "user", "pwd")
	if err != nil {
		fmt.Printf("err %v \n", err.Error())
		return
	}
	defer cli.Close()
	cmds := []CommandInfo{
		{
			Cmd:           "su - root",
			IsInteractive: true,
			EndPoint:      "Password",
		},
		{
			Cmd:           "pwd",
			IsInteractive: false,
			EndPoint:      "",
		},
		{
			Cmd:           "whoami",
			IsInteractive: false,
			EndPoint:      "",
		},
		{
			Cmd:           "cd /home/",
			IsInteractive: false,
			EndPoint:      "",
		},
		{
			Cmd:           "pwd",
			IsInteractive: false,
			EndPoint:      "",
		},
		{
			Cmd:           "tail -f a.log",
			IsInteractive: false,
			EndPoint:      "",
		},
	}
	ret, err := cli.ExecuteWithShell(cmds, 10*time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, r := range ret {
		fmt.Println("cmd:", r.Cmd.Cmd)
		fmt.Println("result:", r.Result)
	}
}
