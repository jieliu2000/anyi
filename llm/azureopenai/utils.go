package azureopenai

import (
	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/jieliu2000/anyi/llm/chat"
)

func ConvertToAzureOpenAIMessageCompletions(messages []chat.Message) []azopenai.ChatRequestMessageClassification {
	result := []azopenai.ChatRequestMessageClassification{}
	for _, msg := range messages {
		switch msg.Role {
		case string(azopenai.ChatRoleUser):
			result = append(result, &azopenai.ChatRequestUserMessage{
				Content: azopenai.NewChatRequestUserMessageContent(msg.Content),
			})
		case string(azopenai.ChatRoleSystem):

			result = append(result, &azopenai.ChatRequestSystemMessage{
				Content: to.Ptr(msg.Content),
			})
		case string(azopenai.ChatRoleAssistant):
			result = append(result, &azopenai.ChatRequestAssistantMessage{
				Content: to.Ptr(msg.Content),
			})
		default:
			continue
		}
	}
	return result

}
