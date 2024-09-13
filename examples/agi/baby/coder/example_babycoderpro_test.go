package babycoder

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
								"template": `Perform one task based on the following objective: {{.Objective}}
Take into account these previously completed tasks: 
{{.Result}}
Your task: {{.Task.Description}}
Response:`,
							},
						},
					},
				},
			},
			{
				Name: "taskInitiator",
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
								"template": `You are an AGI agent responsible for creating a detailed JSON checklist of tasks that will guide other AGI agents to complete a given programming objective. Your task is to analyze the provided objective and generate a well-structured checklist with a clear starting point and end point, as well as tasks broken down to be very specific, clear, and executable by other agents without the context of other tasks.

    The current agents work as follows:
    - code_writer_agent: Writes code snippets or functions and saves them to the appropriate files. This agent can also append code to existing files if required.
    - code_refactor_agent: Responsible for modifying and refactoring existing code to meet the requirements of the task.
    - command_executor_agent: Executes terminal commands for tasks such as creating directories, installing dependencies, etc.

    Keep in mind that the agents cannot open files in text editors, and tasks should be designed to work within these agent capabilities.

    Here is the programming objective you need to create a checklist for: {objective}.

    To generate the checklist, follow these steps:

    1. Analyze the objective to identify the high-level requirements and goals of the project. This will help you understand the scope and create a comprehensive checklist.

    2. Break down the objective into smaller, highly specific tasks that can be worked on independently by other agents. Ensure that the tasks are designed to be executed by the available agents (code_writer_agent, code_refactor and command_executor_agent) without requiring opening files in text editors.

    3. Assign a unique ID to each task for easy tracking and organization. This will help the agents to identify and refer to specific tasks in the checklist.

    4. Organize the tasks in a logical order, with a clear starting point and end point. The starting point should represent the initial setup or groundwork necessary for the project, while the end point should signify the completion of the objective and any finalization steps.

    5. Provide the current context for each task, which should be sufficient for the agents to understand and execute the task without referring to other tasks in the checklist. This will help agents avoid task duplication.

    6. Pay close attention to the objective and make sure the tasks implement all necessary pieces needed to make the program work.
    
    7. Compile the tasks into a well-structured JSON format, ensuring that it is easy to read and parse by other AGI agents. The JSON should include fields such as task ID, description and file_path.

    IMPORTANT: BE VERY CAREFUL WITH IMPORTS AND MANAGING MULTIPLE FILES. REMEMBER EACH AGENT WILL ONLY SEE A SINGLE TASK. ASK YOURSELF WHAT INFORMATION YOU NEED TO INCLUDE IN THE CONTEXT OF EACH TASK TO MAKE SURE THE AGENT CAN EXECUTE THE TASK WITHOUT SEEING THE OTHER TASKS OR WHAT WAS ACCOMPLISHED IN OTHER TASKS.

    Pay attention to the way files are passed in the tasks, always use full paths. For example 'project/main.py'.

    Make sure tasks are not duplicated.

    Do not take long and complex routes, minimize tasks and steps as much as possible.

    Here is a sample JSON output for a checklist:

            {{
                "tasks": [
                    {{
                    "id": 1,
                    "description": "Run a command to create the project directory named 'project'",
                    "file_path": "./project",
                    }},
                    {{
                    "id": 2,
                    "description": "Run a command to Install the following dependencies: 'numpy', 'pandas', 'scikit-learn', 'matplotlib'",
                    "file_path": "null",
                    }},
                    {{
                    "id": 3,
                    "description": "Write code to create a function named 'parser' that takes an input named 'input' of type str, [perform a specific task on it], and returns a specific output",
                    "file_path": "./project/main.py",
                    }},
                    ...
                    {{
                    "id": N,
                    "description": "...",
                    }}
                ],
            }}

    The tasks will be executed by either of the three agents: command_executor, code_writer or code_refactor. They can't interact with programs. They can either run terminal commands or write code snippets. Their output is controlled by other functions to run the commands or save their output to code files. Make sure the tasks are compatible with the current agents. ALL tasks MUST start either with the following phrases: 'Run a command to...', 'Write code to...', 'Edit existing code to...' depending on the agent that will execute the task. RETURN JSON ONLY:							`,
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
{{range $index, $task := .Tasks}}	- {{$task.Description}}
{{end}}
Consider the ultimate objective of your team: {{.Objective}}.

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

func Example_coderTaskAGI() {

	objective := `Create a Python program that consists of a single class named 'TemperatureConverter' in a file named 'temperature_converter.py'. The class should have the following methods:

- celsius_to_fahrenheit(self, celsius: float) -> float: Converts Celsius temperature to Fahrenheit.
- fahrenheit_to_celsius(self, fahrenheit: float) -> float: Converts Fahrenheit temperature to Celsius.

Create a separate 'main.py' file that imports the 'TemperatureConverter' class, takes user input for the temperature value and the unit, converts the temperature to the other unit, and then prints the result.`

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
