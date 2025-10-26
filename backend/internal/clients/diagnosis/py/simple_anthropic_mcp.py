"""
简化版 Anthropic MCP 集成
最少代码实现 Claude + Prometheus MCP
"""

import asyncio
import json
import os
from contextlib import AsyncExitStack
from anthropic import AsyncAnthropic
from mcp.client.session import ClientSession
from mcp.client.stdio import StdioServerParameters, stdio_client


async def simple_diagnosis(
    prompt: str,
    anthropic_api_key: str,
    prometheus_url: str = "http://localhost:9300",
    github_token: str = None,
    enable_github_mcp: bool = True,
    github_toolsets: str = "repos,issues,pull_requests,releases",
    anthropic_base_url: str = "https://api-inference.modelscope.cn",
    model: str = "Qwen/Qwen3-Coder-480B-A35B-Instruct",
):
    """
    使用 MCP 和 AI 进行诊断分析

    Args:
        prompt: 完整的 AI prompt（由 Go 代码构建）
        anthropic_api_key: Anthropic API Key
        prometheus_url: Prometheus 地址
        github_token: GitHub Personal Access Token（可选）
        github_toolsets: GitHub MCP 工具集（默认 "repos,issues,pull_requests,releases"）
        anthropic_base_url: Anthropic API Base URL (可选，用于兼容服务)
        model: 模型名称

    Returns:
        诊断报告文本
    """

    # 使用 AsyncExitStack 管理多个 MCP sessions
    async with AsyncExitStack() as stack:
        # ==================== 1. 启动 Prometheus MCP ====================
        print(f"[Prometheus MCP] 启动中...")
        print(f"[Prometheus MCP] Prometheus URL: {prometheus_url}")

        prometheus_params = StdioServerParameters(
            command="python",
            args=["-m", "prometheus_mcp_server.main"],
            env={
                "PROMETHEUS_URL": prometheus_url,
                "PROMETHEUS_MCP_SERVER_TRANSPORT": "stdio",
            }
        )

        prom_read, prom_write = await stack.enter_async_context(
            stdio_client(prometheus_params)
        )
        prom_session = await stack.enter_async_context(
            ClientSession(prom_read, prom_write)
        )
        await prom_session.initialize()
        print(f"[Prometheus MCP] ✅ 连接成功")

        # ==================== 2. 启动 GitHub MCP（如果启用）====================
        github_session = None
        if enable_github_mcp and github_token:
            print(f"\n[GitHub MCP] 启动中...")
            print(f"[GitHub MCP] 工具集: {github_toolsets}")

            github_params = StdioServerParameters(
                command="github-mcp-server",
                args=["stdio"],
                env={
                    "GITHUB_PERSONAL_ACCESS_TOKEN": github_token,
                    "GITHUB_TOOLSETS": github_toolsets,
                }
            )

            try:
                gh_read, gh_write = await stack.enter_async_context(
                    stdio_client(github_params)
                )
                github_session = await stack.enter_async_context(
                    ClientSession(gh_read, gh_write)
                )
                await github_session.initialize()
                print(f"[GitHub MCP] ✅ 连接成功")
            except Exception as e:
                print(f"[GitHub MCP] ⚠️  启动失败: {e}")
                print(f"[GitHub MCP] 将仅使用 Prometheus MCP 继续")
        elif enable_github_mcp and not github_token:
            print(f"\n[GitHub MCP] ⚠️  未提供 GitHub Token，跳过")

        # ==================== 3. 获取工具列表并创建路由器 ====================
        tool_router = {}  # {tool_name: session}
        all_tools = []

        # Prometheus MCP 工具
        prom_tools = await prom_session.list_tools()
        print(f"\n[Prometheus MCP] 加载了 {len(prom_tools.tools)} 个工具:")
        for tool in prom_tools.tools:
            print(f"  - {tool.name}")
            tool_router[tool.name] = prom_session
            all_tools.append(tool)

        # GitHub MCP 工具（如果可用）
        if github_session:
            gh_tools = await github_session.list_tools()
            print(f"\n[GitHub MCP] 加载了 {len(gh_tools.tools)} 个工具:")
            for tool in gh_tools.tools:
                print(f"  - {tool.name}")
                tool_router[tool.name] = github_session
                all_tools.append(tool)

        print(f"\n[MCP] 总共加载了 {len(all_tools)} 个工具")

        # ==================== 4. 转换为 Anthropic 格式 ====================
        tools = [
            {
                "name": tool.name,
                "description": tool.description,
                "input_schema": tool.inputSchema,
            }
            for tool in all_tools
        ]


        # ==================== 5. 创建 Claude 客户端 ====================
        print(f"\n[AI] 使用模型: {model}")
        if anthropic_base_url:
            print(f"[AI] Base URL: {anthropic_base_url}")
            client = AsyncAnthropic(
                api_key=anthropic_api_key,
                base_url=anthropic_base_url,
            )
        else:
            client = AsyncAnthropic(api_key=anthropic_api_key)

        # ==================== 6. 构建消息 ====================
        messages = [{
            "role": "user",
            "content": prompt
        }]

        # ==================== 7. AI 对话循环 ====================
        print(f"\n[AI] 开始分析...\n")


        for iteration in range(20):  # 最多 20 轮
            print(f"[AI] 第 {iteration + 1} 轮对话")

            response = await client.messages.create(
                model=model,
                max_tokens=4096,
                messages=messages,
                tools=tools,
            )

            print(f"[AI] Stop reason: {response.stop_reason}")

            if response.stop_reason == "end_turn":
                # 完成，提取最终响应文本
                final_text = ""
                for block in response.content:
                    if block.type == "text":
                        final_text += block.text

                print(f"\n[AI] ✅ 分析完成")

                # 直接返回纯文本
                return final_text.strip()

            # 添加 Claude 响应
            messages.append({"role": "assistant", "content": response.content})

            # ==================== 执行工具调用（使用路由器）====================
            tool_results = []
            for block in response.content:
                if block.type == "tool_use":
                    # 使用工具路由器找到对应的 session
                    session = tool_router.get(block.name)

                    if not session:
                        print(f"[MCP] ⚠️  未找到工具: {block.name}")
                        tool_results.append({
                            "type": "tool_result",
                            "tool_use_id": block.id,
                            "content": f"错误：未找到工具 {block.name}",
                            "is_error": True,
                        })
                        continue

                    # 判断是哪个 MCP（用于日志）
                    mcp_name = "Prometheus MCP" if session == prom_session else "GitHub MCP"

                    print(f"[{mcp_name}] 调用工具: {block.name}")
                    print(f"[{mcp_name}] 参数: {json.dumps(block.input, indent=2, ensure_ascii=False)}")

                    try:
                        # 调用 MCP 工具
                        result = await session.call_tool(
                            block.name,
                            block.input
                        )

                        # 提取文本结果
                        result_text = "\n".join([
                            c.text for c in result.content
                            if hasattr(c, "text")
                        ])

                        print(f"[{mcp_name}] 结果: {result_text[:200]}...")

                        tool_results.append({
                            "type": "tool_result",
                            "tool_use_id": block.id,
                            "content": result_text,
                        })

                    except Exception as e:
                        print(f"[{mcp_name}] ❌ 工具调用失败: {e}")
                        tool_results.append({
                            "type": "tool_result",
                            "tool_use_id": block.id,
                            "content": f"工具调用失败: {str(e)}",
                            "is_error": True,
                        })

            # 添加工具结果
            messages.append({"role": "user", "content": tool_results})
            print()

        return "诊断分析达到最大迭代次数，请检查配置或联系技术支持。"


# 使用示例
async def main():
    from dotenv import load_dotenv
    load_dotenv()
    
    # 示例 prompt（仅使用 Prometheus MCP）
    test_prompt_prometheus = """你是一个专业的 DevOps 运维诊断专家。

收到以下告警信息：
- 告警名称: HighCPUUsage
- 实例: localhost:9100

请使用 Prometheus MCP 工具查询相关指标，分析问题原因并生成诊断报告。
"""

    # 示例 prompt（同时使用 Prometheus 和 GitHub MCP）
    test_prompt_github = """你是一个代码审查专家。

请分析 Z3Labs/Hackathon 仓库的最新 release：
1. 获取最新 release
2. 提取其中的 PR 列表
3. 分析每个 PR 的代码变更，查找潜在的 bug
4. 生成详细的 bug 报告

同时检查 Prometheus 监控指标是否正常。
"""

    # 选择测试场景
    use_github = os.getenv("TEST_GITHUB_MCP", "false").lower() == "true"

    result = await simple_diagnosis(
        prompt=test_prompt_github if use_github else test_prompt_prometheus,
        anthropic_api_key=os.getenv("ANTHROPIC_API_KEY"),
        prometheus_url=os.getenv("PROMETHEUS_URL", "http://localhost:9090"),
        github_token=os.getenv("GITHUB_TOKEN"),  # 新增
        enable_github_mcp=use_github,  # 新增
        github_toolsets=os.getenv("GITHUB_TOOLSETS", "repos,issues,pull_requests,releases"),  # 新增
        anthropic_base_url=os.getenv("ANTHROPIC_BASE_URL"),
        model=os.getenv("CLAUDE_MODEL", "claude-3-5-sonnet-20241022"),
    )
    
    print("\n" + "=" * 80)
    print("最终诊断报告")
    print("=" * 80)
    print(result)


if __name__ == "__main__":
    asyncio.run(main())
