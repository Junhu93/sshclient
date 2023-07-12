package sshclient

import (
	"context"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	client *ssh.Client
	user   string
	ip     string
	port   string
}

type CommandInfo struct {
	Cmd           string
	IsInteractive bool
	EndPoint      string
}

type CommandResult struct {
	Cmd    CommandInfo
	Result string
}

func NewClient(ip, port, user, pwd string) (*Client, error) {
	sshcli, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", ip, port), &ssh.ClientConfig{

		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pwd),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		}),
		// Timeout 建立tcp连接的超时时间
		Timeout: time.Duration(10) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	return &Client{
		client: sshcli,
		user:   user,
		ip:     ip,
		port:   port,
	}, nil
}

func (cli *Client) Close() error {
	return cli.client.Close()
}

func (cli *Client) Run(cmd string) error {
	session, err := cli.client.NewSession()
	if err != nil {
		fmt.Printf("client NewSession fialed ,err = %v \n", err.Error())
		return err
	}
	defer session.Close()
	return session.Run(cmd)
}

func (cli *Client) CombineOutput(cmd string) ([]byte, error) {
	session, err := cli.client.NewSession()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	defer session.Close()
	return session.CombinedOutput(cmd)
}

func (cli *Client) ExecuteWithShell(cmds []CommandInfo, timeout time.Duration) ([]CommandResult, error) {
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	s, _, err := CreateShell(ctx, cli)
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				s.session.Close()
				return
			}
		}
	}()
	defer s.Close()

	output, err := s.ExecuteCommands(cmds)
	if err != nil {
		return nil, err
	}
	return output, nil
}
