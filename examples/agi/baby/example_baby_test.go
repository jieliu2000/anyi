package baby

import (
	"log"
	"os"
	"strings"

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
	Tasks     []*TaskData
}

func ExecuteTask(task *TaskData, taskResultList []TaskResult, objective string) TaskResult {
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

	return TaskResult{
		Result: memory.Text,
		Task:   task.Description,
	}
}

func CreateTask(objective string, result TaskResult, task *TaskData, queue *Queue[TaskData]) []*TaskData {
	context := TaskExecutorContext{
		Objective: objective,
		Tasks:     queue.data,
		Result:    result.Result,
		Task:      task,
	}
	executorFlow, _ := anyi.GetFlow("createTask")
	memory := executorFlow.NewShortTermMemory("", context)

	memory, err := executorFlow.Run(*memory)
	if err != nil {
		log.Fatal(err)
	}

	output := memory.Text
	tasks := strings.Split(output, "\n")

	taskDataList := []*TaskData{}

	for _, task := range tasks {
		taskData := &TaskData{
			Description: task,
		}
		taskDataList = append(taskDataList, taskData)
	}
	return taskDataList
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
								"templateFile": "execute_task.tmpl",
							},
						},
					},
				},
			},
			{
				Name: "createTask",
				Steps: []anyi.StepConfig{
					{
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							Config: map[string]interface{}{
								"template": `You are to use the result from an execution agent to create new tasks with the following objective: {{.Objective}}.
The last completed task has the result: {{.Result}}
This result was based on this task description: {{.Task.Description}} {{if .Tasks}}
These are incomplete tasks:{{range $index, $task := .Tasks}}
	- {{$task.Description}}
{{end}}{{end}}
Based on the result, return a list of tasks to be completed in order to meet the objective. 
{{if .Tasks}}These new tasks must not overlap with incomplete tasks. {{end}}
Return one task per line in your response. The result must be a numbered list in the format:

#. First task
#. Second task

The number of each entry must be followed by a period. If your list is empty, write "There are no tasks to add at this time."
Unless your list is empty, do not include any headers before your numbered list or follow your numbered list with any other output.
`,
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

	objective := "Create an HTML web page with good UI for accessing ollama."

	taskQueue := Queue[TaskData]{
		data: []*TaskData{
			{Description: "Find out how to access ollama"},
		},
	}
	resultList := []TaskResult{}

	InitAnyi()
	loop := true

	for loop == true {
		if taskQueue.Len() > 0 {
			log.Println("-------------Task List------------")
			for i, t := range taskQueue.data {
				log.Printf("%d. %s\n", i+1, t.Description)
			}
			task := taskQueue.Poll()
			log.Printf("-------------Next task to do------------\n")
			log.Println(task.Description)
			result := ExecuteTask(task, resultList, objective)
			log.Printf("-------------Task result------------\n")
			log.Println(result)

			resultList = append(resultList, result)

			newTasks := CreateTask(objective, result, task, &taskQueue)

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
	// Output:
	// -------------Task List------------
	// 	1. Create a new project in python
	// -------------Next task to do------------
}
