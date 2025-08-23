package main

import (
	"fmt"
	"log"

	"github.com/jieliu2000/anyi"
)

// 演示去掉 AgentRegistry 后的简化架构
func main() {
	fmt.Println("=== 简化的 Agent 框架演示 ===")
	fmt.Println()

	// 配置内容 - 现在不再需要复杂的适配器
	configContent := `
clients:
  - name: mock-client
    type: openai  # 使用 mock 进行演示
    apiKey: "mock-key"
    default: true

flows:
  - name: simple_flow
    clientName: mock-client
    steps:
      - name: process
        executor:
          type: llm
          withconfig:
            prompt: "Process this request: {{.input}}"

agents:
  - name: simple_agent
    description: "简化的演示 Agent"
    flows: ["simple_flow"]
    clientName: mock-client
    config:
      max_iterations: 1
`

	// 加载配置
	err := anyi.ConfigFromString(configContent, "yaml")
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}

	fmt.Println("✅ 配置加载成功")

	// 获取 Agent
	agent, err := anyi.GetAgent("simple_agent")
	if err != nil {
		log.Fatalf("获取 Agent 失败: %v", err)
	}

	fmt.Printf("✅ 成功获取 Agent: %s\n", agent.GetName())
	fmt.Printf("   描述: %s\n", agent.Description)
	fmt.Printf("   可用流程: %v\n", agent.GetFlows())
	fmt.Printf("   LLM 客户端: %s\n", agent.GetClientName())

	fmt.Println()
	fmt.Println("📝 架构简化总结:")
	fmt.Println("   ❌ 去掉了: AgentRegistry 接口")
	fmt.Println("   ❌ 去掉了: FunctionalRegistryAdapter")
	fmt.Println("   ❌ 去掉了: RegistryFunctions 结构")
	fmt.Println("   ✅ 保留了: 简单的函数参数传递")
	fmt.Println("   ✅ 保留了: 完整的 Agent 功能")
	fmt.Println("   ✅ 保留了: anyi.ConfigFromFile() → anyi.GetAgent() → agent.Execute() 流程")

	fmt.Println()
	fmt.Println("🎯 核心改进:")
	fmt.Println("   • 减少了代码复杂度")
	fmt.Println("   • 消除了不必要的抽象层")
	fmt.Println("   • 保持了相同的用户接口")
	fmt.Println("   • 提高了代码可读性")

	fmt.Println()
	fmt.Println("🎉 Agent 框架简化完成！")
}
