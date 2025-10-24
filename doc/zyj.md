总体任务
1. github action 编包（编译后可在ubuntu执行的二进制文件，加项目 backend/etc/hackathon-api.yaml文件） + 上传kodo
2. ai智能解析发布后的node exporter指标，生成原因分析报告
   1. 机器指标采集使用普罗米修斯系的 node exporter
   2. 当我拿到指标数据后，调用ai接口生成分析报告，需要编写提示词
      - 代码判断是否是异常指标后，确认异常后再发给ai
      - 分为多种异常场景：物理机资源问题，go runtime问题等等待补充（根据node exporter 可判断的异常）
      - 是否有mcp能够让ai主动调用某个接口，我来提供最新的指标数据？待调研
        - 若没有mcp，当ai需要新的数据时，需要定义好提示词限制输出格式
      - ai生成的报告需要包含：异常指标列表，异常原因分析，解决建议
3. mock 不同异常场景的 node exporter 指标数据，验证ai生成报告的准确性