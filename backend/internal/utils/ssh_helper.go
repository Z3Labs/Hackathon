package utils

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// TestSSHConnectionAndGetHostname 测试SSH连接并获取远程机器的hostname
func TestSSHConnectionAndGetHostname(ip string, port int, username, password string) (bool, string, string, error) {
	// 配置SSH客户端
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 仅用于测试
		Timeout:         5 * time.Second,
	}

	// 建立SSH连接
	address := fmt.Sprintf("%s:%d", ip, port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "unable to authenticate") {
			return false, "", fmt.Sprintf("SSH认证失败，请检查用户名或密码是否正确"), err
		}
		if strings.Contains(errMsg, "connection refused") {
			return false, "", fmt.Sprintf("无法连接到 %s:%d，请检查端口是否正确或服务是否已启动", ip, port), err
		}
		if strings.Contains(errMsg, "no route to host") || strings.Contains(errMsg, "timed out") {
			return false, "", fmt.Sprintf("无法连接到 %s:%d，请检查网络连接和IP地址是否正确", ip, port), err
		}
		return false, "", fmt.Sprintf("无法连接到 %s:%d: %v", ip, port, err), err
	}
	defer client.Close()

	// 创建会话
	session, err := client.NewSession()
	if err != nil {
		return false, "", fmt.Sprintf("无法创建SSH会话: %v", err), err
	}
	defer session.Close()

	// 执行 hostname 命令
	output, err := session.CombinedOutput("hostname")
	if err != nil {
		return false, "", fmt.Sprintf("无法获取hostname: %v", err), err
	}

	// 清理hostname输出（去除换行符）
	hostname := strings.TrimSpace(string(output))
	if hostname == "" {
		return false, "", "hostname为空", fmt.Errorf("hostname为空")
	}

	return true, hostname, fmt.Sprintf("成功连接并获取hostname: %s", hostname), nil
}
