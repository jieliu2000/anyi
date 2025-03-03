package llm_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi/llm"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/deepseek"
)

func Example_deepseek() {
	// 请确保已设置DEEPSEEK_API_KEY环境变量
	config := deepseek.DefaultConfig(os.Getenv("DEEPSEEK_API_KEY"), "deepseek-chat")
	client, err := llm.NewClient(config)

	if err != nil {
		log.Fatalf("创建客户端失败: %v", err)
	}

	messages := []chat.Message{
		{Role: "user", Content: "5+1=?"},
	}
	message, _, _ := client.Chat(messages, nil)

	log.Printf("响应结果: %s\n", message.Content)
}
