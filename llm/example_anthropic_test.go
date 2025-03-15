package llm_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/anthropic"
	"github.com/jieliu2000/anyi/llm/chat"
)

func Example_anthropic() {
	// 确保设置ANTHROPIC_API_KEY环境变量
	config := anthropic.DefaultConfig(os.Getenv("ANTHROPIC_API_KEY"))

	// 可以指定特定模型，例如Claude 3 Sonnet
	config.Model = "claude-3-sonnet-20240229"

	client, err := llm.NewClient(config)
	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "你能简要介绍一下你自己吗？"},
	}
	message, _, err := client.Chat(messages, nil)
	if err != nil {
		log.Fatalf("聊天失败: %v", err)
	}

	log.Printf("响应结果: %s\n", message.Content)
}
