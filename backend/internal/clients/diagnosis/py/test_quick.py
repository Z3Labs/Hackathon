"""
快速测试 - 验证 ModelScope API + Prometheus 配置
"""

import asyncio
import os
from dotenv import load_dotenv
from simple_anthropic_mcp import simple_diagnosis

# 加载环境变量
load_dotenv()


async def test_connection():
    """测试 API 连接"""
    print("=" * 80)
    print("配置信息")
    print("=" * 80)
    print(f"API Key: {os.getenv('CUSTOM_ANTHROPIC_API_KEY')}...")
    print(f"Base URL: {os.getenv('CUSTOM_ANTHROPIC_BASE_URL')}")
    print(f"Model: {os.getenv('CUSTOM_CLAUDE_MODEL')}")
    print(f"Prometheus: {os.getenv('PROMETHEUS_URL')}")
    print()
    
    # 构建测试 prompt
    test_prompt = """你是一个专业的 DevOps 运维诊断专家，擅长分析系统告警并定位问题根因。

**收到以下告警信息**：

告警名称: HighCPUUsage
告警状态: firing
严重程度: critical
描述信息: CPU 使用率超过 80% 阈值
触发值: 92.50
实例: localhost:9301
开始时间: 2025-01-15 14:23:15

**你的任务**：

1. **使用 Prometheus MCP 工具查询相关指标**
   - 使用 get_targets() 检查 Prometheus 抓取目标状态
   - 使用 execute_query() 查询关键指标（CPU、内存、网络等）
   - 使用 execute_range_query() 获取时间范围内的趋势数据

2. **分析问题的根本原因**
   - 结合告警信息和查询到的指标数据
   - 分析指标之间的关联关系
   - 识别异常模式和趋势

3. **生成诊断报告**
   请以清晰的文本格式输出诊断报告，包含：
   - 问题概述
   - 根因分析
   - 影响范围
   - 解决方案

**重要提示**：
- 请使用 MCP 工具主动查询所需的指标数据
- 在诊断报告中，只输出分析结果和建议，不需要列出查询到的原始指标数据
- 报告应该简洁明了，便于运维人员快速理解和处理

现在请开始诊断分析："""
    
    print("=" * 80)
    print("开始诊断测试")
    print("=" * 80)
    
    try:
        result = await simple_diagnosis(
            prompt=test_prompt,
            anthropic_api_key=os.getenv('CUSTOM_ANTHROPIC_API_KEY'),
            prometheus_url=os.getenv('PROMETHEUS_URL'),
            anthropic_base_url=os.getenv('CUSTOM_ANTHROPIC_BASE_URL'),
            model=os.getenv('CUSTOM_CLAUDE_MODEL'),
        )
        
        print("\n" + "=" * 80)
        print("诊断结果")
        print("=" * 80)
        print(result)
        
    except Exception as e:
        print(f"\n❌ 测试失败: {e}")
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    asyncio.run(test_connection())
