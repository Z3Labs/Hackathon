package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Z3Labs/Hackathon/backend/internal/logic/deployments/executor"
	"github.com/Z3Labs/Hackathon/backend/internal/model"
)

func main() {
	var (
		action      string
		host        string
		service     string
		version     string
		prevVersion string
		packageURL  string
		sha256      string
		playbookPath string
		timeout     int
	)

	flag.StringVar(&action, "action", "deploy", "操作类型: deploy, rollback, status")
	flag.StringVar(&host, "host", "", "目标主机地址 (必填)")
	flag.StringVar(&service, "service", "", "服务名称 (必填)")
	flag.StringVar(&version, "version", "", "部署版本 (必填)")
	flag.StringVar(&prevVersion, "prev-version", "", "上一个版本 (rollback 时必填)")
	flag.StringVar(&packageURL, "package-url", "", "软件包 URL (必填)")
	flag.StringVar(&sha256, "sha256", "", "软件包 SHA256 (必填)")
	flag.StringVar(&playbookPath, "playbook", "/workspace/backend/playbooks/deploy.yml", "Ansible playbook 路径")
	flag.IntVar(&timeout, "timeout", 600, "执行超时时间（秒）")

	flag.Parse()

	if host == "" || service == "" {
		flag.Usage()
		log.Fatal("错误: host 和 service 参数必填")
	}

	if action == "deploy" || action == "rollback" {
		if version == "" || packageURL == "" || sha256 == "" {
			flag.Usage()
			log.Fatal("错误: deploy 和 rollback 操作需要 version, package-url, sha256 参数")
		}
	}

	if action == "rollback" && prevVersion == "" {
		flag.Usage()
		log.Fatal("错误: rollback 操作需要 prev-version 参数")
	}

	if playbookPath != "" {
		os.Setenv("ANSIBLE_PLAYBOOK_PATH", playbookPath)
	}

	config := executor.ExecutorConfig{
		Platform:    string(model.PlatformPhysical),
		Host:        host,
		Service:     service,
		Version:     version,
		PrevVersion: prevVersion,
		PackageURL:  packageURL,
		SHA256:      sha256,
	}

	exec := executor.NewAnsibleExecutor(config)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	fmt.Printf("=== Ansible Executor 测试工具 ===\n\n")
	fmt.Printf("操作类型: %s\n", action)
	fmt.Printf("目标主机: %s\n", host)
	fmt.Printf("服务名称: %s\n", service)
	fmt.Printf("部署版本: %s\n", version)
	if prevVersion != "" {
		fmt.Printf("上一版本: %s\n", prevVersion)
	}
	fmt.Printf("软件包URL: %s\n", packageURL)
	fmt.Printf("SHA256: %s\n", sha256)
	fmt.Printf("Playbook: %s\n", playbookPath)
	fmt.Printf("\n================================\n\n")

	var err error
	startTime := time.Now()

	switch action {
	case "deploy":
		fmt.Println("开始执行部署...")
		err = exec.Deploy(ctx)
	case "rollback":
		fmt.Println("开始执行回滚...")
		err = exec.Rollback(ctx)
	case "status":
		fmt.Println("查询部署状态...")
		status, statusErr := exec.GetStatus(ctx)
		if statusErr != nil {
			log.Fatalf("获取状态失败: %v", statusErr)
		}
		printStatus(status)
		return
	default:
		log.Fatalf("未知操作类型: %s", action)
	}

	duration := time.Since(startTime)

	if err != nil {
		fmt.Printf("\n❌ 操作失败 (耗时: %v)\n", duration)
		log.Fatalf("错误详情: %v", err)
	}

	fmt.Printf("\n✅ 操作成功 (耗时: %v)\n", duration)

	status, _ := exec.GetStatus(ctx)
	fmt.Println("\n当前状态:")
	printStatus(status)
}

func printStatus(status *model.NodeDeployStatusRecord) {
	fmt.Printf("  主机: %s\n", status.Host)
	fmt.Printf("  服务: %s\n", status.Service)
	fmt.Printf("  当前版本: %s\n", status.CurrentVersion)
	fmt.Printf("  部署中版本: %s\n", status.DeployingVersion)
	fmt.Printf("  上一版本: %s\n", status.PrevVersion)
	fmt.Printf("  平台: %s\n", status.Platform)
	fmt.Printf("  状态: %s\n", status.State)
	if status.LastError != "" {
		fmt.Printf("  最后错误: %s\n", status.LastError)
	}
	fmt.Printf("  更新时间: %s\n", status.UpdatedAt.Format("2006-01-02 15:04:05"))
}
