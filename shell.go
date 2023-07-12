package sshclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type Shell struct {
	session    *ssh.Session
	ctx        context.Context
	inPipe     io.WriteCloser
	outPipe    io.Reader
	outputChan chan string
}

func CreateShell(ctx context.Context, cli *Client) (*Shell, string, error) {
	session, err := cli.client.NewSession()
	if err != nil {
		return nil, "", err
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	err = session.RequestPty("xterm", 40, 80, modes)
	if err != nil {
		fmt.Println("RequestPty failed ,err = ", err.Error())
		return nil, "", err
	}
	inPipe, err := session.StdinPipe()
	if err != nil {
		fmt.Printf("error %v  \n", err.Error())
		return nil, "", err
	}
	outPipe, err := session.StdoutPipe()
	if err != nil {
		fmt.Printf("error %v  \n", err.Error())
		return nil, "", err
	}
	outBuf := bytes.NewBuffer(make([]byte, 0))
	session.Stdout = outBuf
	if err := session.Shell(); err != nil {
		fmt.Println("shell failed ,err = ", err.Error())
		return nil, "", err
	}
	shell := &Shell{
		ctx:     ctx,
		session: session,
		inPipe:  inPipe,
		outPipe: outPipe,
	}
	// 添加管道
	shell.outputChan = make(chan string, 0)
	go func(ch chan string) {
		for {
			output, ok := <-ch
			if !ok {
				return
			}
			fmt.Println(output)
		}
	}(shell.outputChan)
	var prompt = "$"
	if cli.user == "root" {
		prompt = "#"
	}

	output, err := shell.ReadCmdOutputByte(true, prompt)
	if err != nil {
		fmt.Println("read output error :", err)
		return nil, "", err
	}
	return shell, output, nil
}

func (s *Shell) Close() error {
	if s.outputChan != nil {
		close(s.outputChan)
	}
	return s.session.Close()
}

func (s *Shell) ReadCmdOutputByte(IsInteractive bool, EndPrompt string) (string, error) {
	output := ""
	for {
		select {
		case <-s.ctx.Done():
			return "", s.ctx.Err()
		default:
			byt := make([]byte, 1024)
			n, err := s.outPipe.Read(byt)
			if err != nil && err != io.EOF {
				return "", err
			}
			if n == 0 {
				continue
			}
			line := string(byt)
			if s.outputChan != nil {
				s.outputChan <- line
			}
			output += line
			if IsInteractive && EndPrompt != "" {
				if strings.Contains(output, EndPrompt) {
					return output, nil
				}
			} else {
				if strings.Contains(output, "$") || strings.Contains(output, "#") {
					return output, nil
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (s *Shell) execute(cmd CommandInfo) (string, error) {
	if cmd.Cmd == "" {
		fmt.Println("empty cmd")
		return "", nil
	}

	_, err := s.inPipe.Write([]byte(cmd.Cmd + "\n"))
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	output, err := s.ReadCmdOutputByte(cmd.IsInteractive, cmd.EndPoint)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return output, nil
}

func (s *Shell) ExecuteCommands(cmds []CommandInfo) ([]CommandResult, error) {
	out := make([]CommandResult, 0, len(cmds))
	for _, cmd := range cmds {
		select {
		case <-s.ctx.Done():
			return nil, s.ctx.Err()
		default:
			output, err := s.execute(cmd)
			if err != nil {
				fmt.Println(err)
				return nil, err
			}
			ret := CommandResult{
				Cmd:    cmd,
				Result: output,
			}
			out = append(out, ret)
		}
	}

	return out, nil
}
