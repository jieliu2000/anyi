package anyi_test

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm"
)

func Example_flowWithDynamicConfig() {
	config := anyi.AnyiConfig{
		Clients: []llm.ClientConfig{
			{
				Name: "dashscope",
				Type: "dashscope",
				Config: map[string]interface{}{
					"model":  "qwen-max",
					"apiKey": os.Getenv("DASHSCOPE_API_KEY"),
				},
			},
		},
		Flows: []anyi.FlowConfig{
			{
				Name: "smart_writer",
				Steps: []anyi.StepConfig{
					{
						Name: "write_story",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": "Write a sci-fi story about {{.Text}}",
							},
						},
					},

					{
						Name: "translate_story",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": `Translate below text to French without any extra output. The text to be translated: 
								'''{{.Text}}'''`,
							},
						},
					},
				},
			},
		},
	}

	anyi.Config(&config)
	flow, err := anyi.GetFlow("smart_writer")
	if err != nil {
		panic(err)
	}
	context, err := flow.RunWithInput("the moon")
	if err != nil {
		panic(err)
	}
	log.Printf("%s", context.Text)
}
