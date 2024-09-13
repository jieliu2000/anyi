package baby

import (
	"fmt"
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
	resultToShow := []TaskResult{}
	for index := range resultList {
		if index > 9 {
			break
		}
		reverseIndex := sliceLen - 1 - index
		r := resultList[reverseIndex]
		resultToShow = append(resultToShow, r)
	}
	for _, r := range resultToShow {
		result += fmt.Sprintf("*. Task: {%s}. Result: {%s}\n", r.Task, r.Result)
	}
	return result
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
	flowContext := executorFlow.NewFlowContext("", context)

	flowContext, err := executorFlow.Run(*flowContext)
	if err != nil {
		log.Fatal(err)
	}

	return TaskResult{
		Result: flowContext.Text,
		Task:   task.Description,
	}
}

func CreateTask(objective string, results []TaskResult, task *TaskData, queue *Queue[TaskData]) []*TaskData {
	completedTask := ""
	for _, result := range results {
		completedTask += fmt.Sprintf(" - %s\n", result.Task)
	}

	context := TaskExecutorContext{
		Objective: objective,
		Tasks:     queue.data,
		Result:    completedTask,
		Task:      task,
	}
	executorFlow, _ := anyi.GetFlow("createTask")
	flowContext := executorFlow.NewFlowContext("", context)

	flowContext, err := executorFlow.Run(*flowContext)
	if err != nil {
		log.Fatal(err)
	}

	output := flowContext.Text
	if output == "notask" {
		return []*TaskData{}
	}

	tasks := strings.Split(output, "\n")

	taskDataList := []*TaskData{}

	for _, task := range tasks {

		task = strings.TrimSpace(task)

		if len(task) == 0 {
			continue
		}
		if !strings.HasPrefix(task, "- ") {
			continue
		}

		task = strings.TrimPrefix(task, "- ")

		taskData := &TaskData{
			Description: task,
		}
		taskDataList = append(taskDataList, taskData)
	}
	return taskDataList
}

func PrioritizeTaskQueue(objective string, result []TaskResult, task *TaskData, queue *Queue[TaskData]) []*TaskData {
	resultString := ResultList(result)
	context := TaskExecutorContext{
		Objective: objective,
		Tasks:     queue.data,
		Result:    resultString,
		Task:      task,
	}
	executorFlow, _ := anyi.GetFlow("prioritizeTask")
	flowContext := executorFlow.NewFlowContext("", context)

	flowContext, err := executorFlow.Run(*flowContext)
	if err != nil {
		log.Fatal(err)
	}

	output := flowContext.Text
	tasks := strings.Split(output, "\n")

	taskDataList := []*TaskData{}
	for _, task := range tasks {
		task = strings.TrimSpace(task)
		if len(task) == 0 {
			continue
		}
		if !strings.HasPrefix(task, "- ") {
			continue
		}

		task = strings.TrimPrefix(task, "- ")

		taskData := &TaskData{
			Description: task,
		}
		taskDataList = append(taskDataList, taskData)
	}

	return taskDataList

}

func InitAnyi() {
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
				Name: "executorTask",
				Steps: []anyi.StepConfig{
					{
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": `Perform one task based on the following objective: {{.Memory.Objective}}
Take into account these previously completed tasks: 
{{.Memory.Result}}
Your task: {{.Memory.Task.Description}}
Response:`,
							},
						},
					},
				},
			},
			{
				Name: "createTask",
				Steps: []anyi.StepConfig{
					{
						Validator: &anyi.ValidatorConfig{
							Type: "string",
							WithConfig: map[string]interface{}{
								"matchRegex": `(^\s*\-\s.*)|(^notask$)`,
							},
						},
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": `You are to use the result from an execution agent to create new tasks with the following objective: {{.Memory.Objective}}.
These are completed tasks: 
{{.Memory.Result}}

{{if .Memory.Tasks}}
These are incomplete tasks:{{range $index, $task := .Memory.Tasks}}
	- {{$task.Description}}
{{end}}{{end}}
Think about the existing completed tasks and the incomplete tasks. If they are not enough to archive the objective, then you should create new tasks and output them.
These new tasks must not overlap with existing completed or incompleted tasks. 
Be very careful when creating new tasks. Review the existing tasks and only create new tasks when the existing tasks cannot achieve the objective.
Return one task per line in your response. The result must be an unordered bullet list in the format:

- First task
- Second task

Use - as bullet symbol. Don't add any numbers or other symbols on each line.
If your list is empty, output "notask" without any other text.
Unless your list is empty, do not include any headers before your bullet list or follow your bullet list with any other output.
`,
							},
						},
					},
				},
			},
			{
				Name: "prioritizeTask",
				Steps: []anyi.StepConfig{
					{
						Validator: &anyi.ValidatorConfig{
							Type: "string",
							WithConfig: map[string]interface{}{
								"matchRegex": `(^\s*\-\s.*)|(^notask$)`,
							},
						},
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": `You are tasked with prioritizing the following tasks: 
{{range $index, $task := .Memory.Tasks}}	- {{$task.Description}}
{{end}}
Consider the ultimate objective of your team: {{.Memory.Objective}}.

Tasks should be sorted from highest to lowest priority, where higher-priority tasks are those that act as pre-requisites or are more essential for meeting the objective.
Output one task per line in your response. The result must be an unordered bullet list in the format:

- First task.
- Second task.

Use - as bullet symbol. Don't add any numbers or other symbols on each line.
Do not include any headers before your ranked list or follow your list with any other output.`,
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

	objective := "Write a python function to add two integers"

	taskQueue := Queue[TaskData]{
		data: []*TaskData{
			{Description: "Write code"},
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
			log.Println(result.Result)

			resultList = append(resultList, result)

			newTasks := CreateTask(objective, resultList, task, &taskQueue)

			if len(newTasks) > 0 {
				for _, nt := range newTasks {
					taskQueue.Add(nt)
				}
			}
			if taskQueue.Len() > 0 {
				prioritizedTasks := PrioritizeTaskQueue(objective, resultList, task, &taskQueue)

				taskQueue.data = prioritizedTasks
			}
		} else {
			log.Println("All tasks completed!")
			loop = false
		}
	}
}
