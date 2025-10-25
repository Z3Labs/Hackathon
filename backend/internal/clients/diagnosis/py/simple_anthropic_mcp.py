"""
简化版 Anthropic MCP 集成
最少代码实现 Claude + Prometheus MCP
"""

import asyncio
import json
import os
from anthropic import AsyncAnthropic
from mcp.client.session import ClientSession
from mcp.client.stdio import StdioServerParameters, stdio_client


async def simple_diagnosis(
    prompt: str,
    anthropic_api_key: str,
    prometheus_url: str = "http://localhost:9300",
    anthropic_base_url: str = "https://api-inference.modelscope.cn",
    model: str = "Qwen/Qwen3-Coder-480B-A35B-Instruct",
):
    """
    使用 MCP 和 AI 进行诊断分析
    
    Args:
        prompt: 完整的 AI prompt（由 Go 代码构建）
        anthropic_api_key: Anthropic API Key
        prometheus_url: Prometheus 地址
        anthropic_base_url: Anthropic API Base URL (可选，用于兼容服务)
        model: 模型名称
        
    Returns:
        诊断报告文本
    """
    
    # 1. 启动 MCP Server
    print(f"[MCP] 启动 Prometheus MCP Server...")
    print(f"[MCP] Prometheus URL: {prometheus_url}")
    
    server_params = StdioServerParameters(
        command="docker",
        args=[
            "run", "-i", "--rm",
            "-e", f"PROMETHEUS_URL={prometheus_url}",
            "ghcr.io/pab1it0/prometheus-mcp-server:latest",
        ],
    )
    
    async with stdio_client(server_params) as (read, write):
        async with ClientSession(read, write) as mcp_session:
            # 2. 初始化 MCP
            await mcp_session.initialize()
            print(f"[MCP] ✅ 连接成功")
            
            # 3. 获取工具列表
            mcp_tools = await mcp_session.list_tools()
            print(f"[MCP] 加载了 {len(mcp_tools.tools)} 个工具:")
            for tool in mcp_tools.tools:
                print(f"  - {tool.name}")
            
            # 4. 转换为 Anthropic 格式
            tools = [
                {
                    "name": tool.name,
                    "description": tool.description,
                    "input_schema": tool.inputSchema,
                }
                for tool in mcp_tools.tools
            ]
            
            # 5. 创建 Claude 客户端
            print(f"\n[AI] 使用模型: {model}")
            if anthropic_base_url:
                print(f"[AI] Base URL: {anthropic_base_url}")
                client = AsyncAnthropic(
                    api_key=anthropic_api_key,
                    base_url=anthropic_base_url,
                )
            else:
                client = AsyncAnthropic(api_key=anthropic_api_key)
            
            # 6. 构建消息（使用外部传入的 prompt）
            messages = [{
                "role": "user",
                "content": prompt
            }]
            
            # 7. 对话循环
            print(f"\n[AI] 开始分析...\n")
            
            for iteration in range(15):  # 最多 15 轮
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
                
                # 执行工具
                tool_results = []
                for block in response.content:
                    if block.type == "tool_use":
                        print(f"[MCP] 调用工具: {block.name}")
                        print(f"[MCP] 参数: {json.dumps(block.input, indent=2)}")
                        
                        # 调用 MCP 工具
                        result = await mcp_session.call_tool(
                            block.name,
                            block.input
                        )
                        
                        # 提取文本结果
                        result_text = "\n".join([
                            c.text for c in result.content 
                            if hasattr(c, "text")
                        ])
                        
                        print(f"[MCP] 结果: {result_text[:200]}...")
                        
                        tool_results.append({
                            "type": "tool_result",
                            "tool_use_id": block.id,
                            "content": result_text,
                        })
                
                # 添加工具结果
                messages.append({"role": "user", "content": tool_results})
                print()
            
            return "诊断分析达到最大迭代次数，请检查配置或联系技术支持。"


# 使用示例
async def main():
    from dotenv import load_dotenv
    load_dotenv()
    
    # 示例 prompt
    test_prompt = """你是一个专业的 DevOps 运维诊断专家。

收到以下告警信息：
- 告警名称: HighCPUUsage
- 实例: localhost:9100

请使用 Prometheus MCP 工具查询相关指标，分析问题原因并生成诊断报告。
"""
    
    result = await simple_diagnosis(
        prompt=test_prompt,
        anthropic_api_key=os.getenv("ANTHROPIC_API_KEY"),
        prometheus_url=os.getenv("PROMETHEUS_URL", "http://localhost:9090"),
        anthropic_base_url=os.getenv("ANTHROPIC_BASE_URL"),
        model=os.getenv("CLAUDE_MODEL", "claude-3-5-sonnet-20241022"),
    )
    
    print("\n" + "=" * 80)
    print("最终诊断报告")
    print("=" * 80)
    print(result)


if __name__ == "__main__":
    asyncio.run(main())
