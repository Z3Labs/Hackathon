#!/usr/bin/env python3
"""
Go 后端调用的 Python MCP 诊断脚本
接收命令行参数，调用 MCP 进行智能诊断，输出诊断报告文本
"""

import sys
import os
import asyncio
import argparse

# 添加当前目录到 Python 路径
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

from simple_anthropic_mcp import simple_diagnosis


async def main():
    """主函数"""
    parser = argparse.ArgumentParser(description='智能诊断系统 - MCP 版本')
    parser.add_argument('--prompt', required=True, help='完整的 AI prompt')
    parser.add_argument('--api-key', required=True, help='AI API Key')
    parser.add_argument('--base-url', required=True, help='AI Base URL')
    parser.add_argument('--model', required=True, help='模型名称')
    parser.add_argument('--prometheus-url', required=True, help='Prometheus URL')

    args = parser.parse_args()

    try:
        # 调用诊断函数
        result = await simple_diagnosis(
            prompt=args.prompt,
            anthropic_api_key=args.api_key,
            prometheus_url=args.prometheus_url,
            anthropic_base_url=args.base_url,
            model=args.model,
        )

        # 输出纯文本到 stdout
        print(result)
        sys.exit(0)

    except Exception as e:
        # 错误信息输出为文本
        error_msg = f"""诊断失败

【错误信息】
{str(e)}

【建议】
1. 检查 Prometheus 连接配置
2. 验证 AI API 配置是否正确
3. 确认网络连接正常
4. 查看详细日志获取更多信息
"""
        print(error_msg)
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
