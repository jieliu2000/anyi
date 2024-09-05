package agi

import (
	"log"
	"os"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/llm"
)

type TaskData struct {
	Description string
}

type TaskResult struct {
	Result string
	Task   string
}

type Queue[T any] struct {
	data []*T
}

func (q *Queue[T]) Add(data *T) {
	q.data = append(q.data, data)
}

func (q *Queue[T]) Len() int {
	return len(q.data)
}

func (q *Queue[T]) IsEmpty() bool {
	return q.Len() == 0
}

func (q *Queue[T]) Poll() *T {
	if len(q.data) > 0 {
		data := q.data[0]
		q.data = q.data[1:]

		return data
	} else {
		return nil
	}
}

func ResultList(resultList []TaskResult) string {
	var result string
	sliceLen := len(resultList)
	for index := range resultList {
		if index > 4 {
			break
		}
		reverseIndex := sliceLen - 1 - index
		r := resultList[reverseIndex]
		result += "\tTask: [[" + r.Task + "]], Result: [[" + r.Result + "]]\n"
	}
	return result + "\n"
}

type TaskExecutorContext struct {
	Objective string
	Result    string
	Task      *TaskData
}

func ExecuteTask(task *TaskData, taskResultList []TaskResult, objective string) string {
	contextPrompt := ResultList(taskResultList)
	context := TaskExecutorContext{
		Objective: objective,
		Result:    contextPrompt,
		Task:      task,
	}
	executorFlow, _ := anyi.GetFlow("executorTask")
	memory := executorFlow.NewShortTermMemory("", context)

	memory, err := executorFlow.Run(*memory)
	if err != nil {
		log.Fatal(err)
	}
	return memory.Text
}

func CreateTask(objective string, result string, taskDescription string, queue *Queue[TaskData]) []*TaskData {
	return nil
}

func PrioritizeTaskQueue(queue *Queue[TaskData]) *Queue[TaskData] {
	return nil
}

func InitAnyi() {
	config := anyi.AnyiConfig{
		Clients: []llm.ClientConfig{
			{
				Name: "zhipu",
				Type: "zhipu",
				Config: map[string]interface{}{
					"apiKey": os.Getenv("ZHIPU_API_KEY"),
					"model":  "glm-4-flash",
				},
			},
		},

		Flows: []anyi.FlowConfig{
			{
				Name: "executorTask",
				Steps: []anyi.StepConfig{
					{
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							Config: map[string]interface{}{
								"template": `Perform one task based on the following objective: {{.Objective}}
Take into account these previously completed tasks: 
{{.Result}}
Your task: {{.Task}}
Response:`,
							},
						},
					},
				},
			},
		},
	}
	anyi.Config(&config)
}

func Example_taskAGI() {

	objective := "Use python to create an AI digital employee project which can generate code for Quasar hybrid mobile app based on user input requirements."

	taskQueue := Queue[TaskData]{
		data: []*TaskData{
			{Description: "Create a new project in python"},
		},
	}
	resultList := []TaskResult{}

	InitAnyi()
	loop := true

	for loop == true {
		if taskQueue.Len() > 0 {
			log.Println("-------------Task List------------")
			for i, t := range taskQueue.data {
				log.Printf("\t%d. %s\n", i+1, t.Description)
			}
			task := taskQueue.Poll()
			log.Printf("-------------Next task to do------------\n")
			log.Println(task.Description)
			result := ExecuteTask(task, resultList, objective)
			log.Printf("-------------Task result------------\n")
			log.Println(result)

			resultList = append(resultList, TaskResult{
				Result: result,
				Task:   task.Description,
			})

			newTasks := CreateTask(objective, result, task.Description, &taskQueue)

			for _, t := range newTasks {
				taskQueue.Add(t)
			}

			prioritizedTaskQueue := PrioritizeTaskQueue(&taskQueue)
			if prioritizedTaskQueue.Len() > 0 {
				taskQueue = *prioritizedTaskQueue
			}

		} else {
			log.Println("All tasks completed!")
			loop = false
		}

	}
}
