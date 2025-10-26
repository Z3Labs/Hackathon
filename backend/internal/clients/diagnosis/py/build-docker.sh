#!/bin/bash
# AI 诊断服务 Docker 镜像构建脚本

set -e  # 遇到错误立即退出

# 颜色输出
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 镜像信息
IMAGE_NAME="diagnosis-service"
IMAGE_TAG="${1:-latest}"  # 默认 tag 为 latest，可通过第一个参数指定
PLATFORM="${2:-}"         # 可选的平台参数（如 linux/amd64）
FULL_IMAGE_NAME="${IMAGE_NAME}:${IMAGE_TAG}"

docker rm -f diagnosis-service
docker rmi diagnosis-service

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}构建 AI 诊断服务 Docker 镜像${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}镜像名称: ${FULL_IMAGE_NAME}${NC}"
if [ -n "$PLATFORM" ]; then
    echo -e "${GREEN}目标平台: ${PLATFORM}${NC}"
fi
echo ""

# 获取脚本所在目录
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# 检查必要文件是否存在
echo -e "${BLUE}检查文件...${NC}"
for file in Dockerfile requirements.txt simple_anthropic_mcp.py diagnosis_runner.py; do
    if [ ! -f "$file" ]; then
        echo -e "${RED}错误: 文件 $file 不存在${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ $file${NC}"
done

# 构建镜像
echo ""
echo -e "${BLUE}开始构建镜像...${NC}"
if [ -n "$PLATFORM" ]; then
    docker build --platform "$PLATFORM" -t "$FULL_IMAGE_NAME" .
else
    docker build -t "$FULL_IMAGE_NAME" .
fi

# 检查构建结果
if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}✓ 镜像构建成功!${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}镜像名称: ${FULL_IMAGE_NAME}${NC}"
    
    # 保存镜像为压缩包
    OUTPUT_FILE="${IMAGE_NAME}-${IMAGE_TAG}.tar.gz"
    echo ""
    echo -e "${BLUE}正在保存镜像为压缩包...${NC}"
    echo -e "${BLUE}输出文件: ${OUTPUT_FILE}${NC}"
    
    docker save "$FULL_IMAGE_NAME" | gzip > "$OUTPUT_FILE"
    
    if [ $? -eq 0 ]; then
        FILE_SIZE=$(du -h "$OUTPUT_FILE" | cut -f1)
        echo -e "${GREEN}✓ 镜像已保存为压缩包!${NC}"
        echo -e "${GREEN}文件路径: $(pwd)/${OUTPUT_FILE}${NC}"
        echo -e "${GREEN}文件大小: ${FILE_SIZE}${NC}"
        echo ""
        echo -e "${BLUE}在其他机器上加载镜像:${NC}"
        echo -e "  ${GREEN}docker load < ${OUTPUT_FILE}${NC}"
        echo -e "  或者: ${GREEN}gunzip -c ${OUTPUT_FILE} | docker load${NC}"
    else
        echo -e "${RED}✗ 保存镜像失败${NC}"
        exit 1
    fi
    
    echo ""
    echo -e "${BLUE}构建参数说明:${NC}"
    echo -e "  默认构建: ${GREEN}./build-docker.sh${NC}"
    echo -e "  指定标签: ${GREEN}./build-docker.sh v1.0${NC}"
    echo -e "  跨平台构建: ${GREEN}./build-docker.sh latest linux/amd64${NC}"
    echo ""
    echo -e "${BLUE}使用方法:${NC}"
    echo -e "  启动容器: ${GREEN}docker run -d --name diagnosis-service ${FULL_IMAGE_NAME}${NC}"
    echo -e "  测试调用: ${GREEN}docker exec -i diagnosis-service python /app/diagnosis_runner.py --help${NC}"
    echo -e "  查看日志: ${GREEN}docker logs diagnosis-service${NC}"
    echo -e "  停止容器: ${GREEN}docker stop diagnosis-service${NC}"
    echo -e "  删除容器: ${GREEN}docker rm diagnosis-service${NC}"
else
    echo ""
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}✗ 镜像构建失败${NC}"
    echo -e "${RED}========================================${NC}"
    exit 1
fi
