#!/bin/bash

set -e

echo "=== Ansible Executor 工具使用示例 ==="
echo ""

if [ ! -f "./ansible-executor-tool" ]; then
    echo "工具未编译，正在编译..."
    go build -o ansible-executor-tool
    echo "编译完成！"
    echo ""
fi

echo "请选择操作类型："
echo "1. 部署 (deploy)"
echo "2. 回滚 (rollback)"
echo "3. 查看状态 (status)"
echo ""
read -p "请输入选项 (1/2/3): " choice

read -p "目标主机 (例如: 192.168.1.100): " host
read -p "服务名称 (例如: my-service): " service

case $choice in
    1)
        echo ""
        echo "=== 部署操作 ==="
        read -p "部署版本 (例如: v1.2.0): " version
        read -p "上一版本 (例如: v1.1.0): " prev_version
        read -p "软件包URL: " package_url
        read -p "SHA256: " sha256
        read -p "Playbook路径 (回车使用默认): " playbook
        
        if [ -z "$playbook" ]; then
            playbook="/workspace/backend/playbooks/deploy.yml"
        fi
        
        echo ""
        echo "执行命令:"
        cmd="./ansible-executor-tool -action=deploy -host=$host -service=$service -version=$version -prev-version=$prev_version -package-url=$package_url -sha256=$sha256 -playbook=$playbook"
        echo "$cmd"
        echo ""
        read -p "确认执行? (y/n): " confirm
        
        if [ "$confirm" = "y" ]; then
            $cmd
        else
            echo "已取消"
        fi
        ;;
        
    2)
        echo ""
        echo "=== 回滚操作 ==="
        read -p "当前版本 (例如: v1.2.0): " version
        read -p "回滚到版本 (例如: v1.1.0): " prev_version
        read -p "软件包URL: " package_url
        read -p "SHA256: " sha256
        
        echo ""
        echo "执行命令:"
        cmd="./ansible-executor-tool -action=rollback -host=$host -service=$service -version=$version -prev-version=$prev_version -package-url=$package_url -sha256=$sha256"
        echo "$cmd"
        echo ""
        read -p "确认执行? (y/n): " confirm
        
        if [ "$confirm" = "y" ]; then
            $cmd
        else
            echo "已取消"
        fi
        ;;
        
    3)
        echo ""
        echo "=== 查看状态 ==="
        echo ""
        echo "执行命令:"
        cmd="./ansible-executor-tool -action=status -host=$host -service=$service"
        echo "$cmd"
        echo ""
        $cmd
        ;;
        
    *)
        echo "无效选项"
        exit 1
        ;;
esac
