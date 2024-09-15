package coder

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm"
)

type TaskData struct {
	Id          int    `json:"id"`
	Description string `json:"description"`

	FilePath string `json:"file_path"`

	IsolatedContext string `json:"isolated_context"`
}

type TaskResult struct {
	Result string
	Task   string
}

type Task struct {
	Objective    string
	Instructions []string
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
				Name: "azureopenai",
				Type: "azureopenai",
				Config: map[string]interface{}{
					"apiKey":            os.Getenv("AZ_OPENAI_API_KEY"),
					"modelDeploymentId": os.Getenv("AZ_OPENAI_MODEL_DEPLOYMENT_ID"),
					"endpoint":          os.Getenv("AZ_OPENAI_ENDPOINT"),
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
								"outputJSON": true,
							},
						},
					},
				},
			},
			{
				Name: "taskInitiator",
				Steps: []anyi.StepConfig{
					{
						Name: "create_initial_task",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": `You are an AGI agent responsible for creating a detailed JSON checklist of tasks that will guide other AGI agents to complete a given programming objective. Your task is to analyze the provided objective and generate a well-structured checklist with a clear starting point and end point, as well as tasks broken down to be very specific, clear, and executable by other agents without the context of other tasks.

    The current agents work as follows:
    - code_writer_agent: Writes code snippets or functions and saves them to the appropriate files. This agent can also append code to existing files if required.
    - code_refactor_agent: Responsible for modifying and refactoring existing code to meet the requirements of the task.
    - command_executor_agent: Executes terminal commands for tasks such as creating directories, installing dependencies, etc.

    Keep in mind that the agents cannot open files in text editors, and tasks should be designed to work within these agent capabilities.

    Here is the programming objective you need to create a checklist for: {{.Memory.Objective}}.

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

            {
                "tasks": [
                    {
                    "id": 1,
                    "description": "Run a command to create the project directory named 'project'",
                    "file_path": "./project",
                    },
                    {
                    "id": 2,
                    "description": "Run a command to Install the following dependencies: 'numpy', 'pandas', 'scikit-learn', 'matplotlib'",
                    "file_path": "null",
                    },
                    {
                    "id": 3,
                    "description": "Write code to create a function named 'parser' that takes an input named 'input' of type str, [perform a specific task on it], and returns a specific output",
                    "file_path": "./project/main.py",
                    },
                    ...
                    {
                    "id": N,
                    "description": "...",
                    }
                ],
            }

    The tasks will be executed by either of the three agents: command_executor, code_writer or code_refactor. They can't interact with programs. They can either run terminal commands or write code snippets. Their output is controlled by other functions to run the commands or save their output to code files. Make sure the tasks are compatible with the current agents. ALL tasks MUST start either with the following phrases: 'Run a command to...', 'Write code to...', 'Edit existing code to...' depending on the agent that will execute the task. 
	Don't append any other characters except the JSON body.
	`,
								"outputJSON": true,
							},
						},
					},
					{
						Name: "tasks_refactor_agent",
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": `You are an AGI tasks_refactor_agent responsible for adapting a task list generated by another agent to ensure the tasks are compatible with the current AGI agents. Your goal is to analyze the task list and make necessary modifications so that the tasks can be executed by the agents listed below

    YOU SHOULD OUTPUT THE MODIFIED TASK LIST IN THE SAME JSON FORMAT AS THE INITIAL TASK LIST. DO NOT CHANGE THE FORMAT OF THE JSON OUTPUT. DO NOT WRITE ANYTHING OTHER THAN THE MODIFIED TASK LIST IN THE JSON FORMAT.
    
    The current agents work as follows:
    - code_writer_agent: Writes code snippets or functions and saves them to the appropriate files. This agent can also append code to existing files if required.
    - code_refactor_agent: Responsible for editing current existing code/files.
    - command_executor_agent: Executes terminal commands for tasks such as creating directories, installing dependencies, etc.

    Here is the overall objective you need to refactor the tasks for: {{.Memory.Objective}}.
    Here is the JSON task list you need to refactor for compatibility with the current agents: {{.Text}}.

    To refactor the task list, follow these steps:
    1. Modify the task descriptions to make them compatible with the current agents, ensuring that the tasks are self-contained, clear, and executable by the agents without additional context. You don't need to mention the agents in the task descriptions, but the tasks should be compatible with the current agents.
    2. If necessary, add new tasks or remove irrelevant tasks to make the task list more suitable for the current agents.
    3. Keep the JSON structure of the task list intact, maintaining the "id", "description" and "file_path" fields for each task.
    4. Pay close attention to the objective and make sure the tasks implement all necessary pieces needed to make the program work.

    Always specify file paths to files. Make sure tasks are not duplicated. Never write code to create files. If needed, use commands to create files and folders.
    Return the updated JSON task list with the following format:

            {
                "tasks": [
                    {
                    "id": 1,
                    "description": "Run a commmand to create a folder named 'project' in the current directory",
                    "file_path": "./project",
                    },
                    {
                    "id": 2,
                    "description": "Write code to print 'Hello World!' with Python",
					"file_path": "./project/main.py",
                    },
                    {
                    "id": 3,
                    "description": "Write code to create a function named 'parser' that takes an input named 'input' of type str, [perform a specific task on it], and returns a specific output",
					"file_path": "./project/main.py",
                    }
                    {
                    "id": 3,
                    "description": "Run a command calling the script in ./project/main.py",
             		"file_path": "./project/main.py",
                    }
                    ...
                ],
            }

    IMPORTANT: All tasks should start either with the following phrases: 'Run a command to...', 'Write a code to...', 'Edit the code to...' depending on the agent that will execute the task:
            
    ALWAYS ENSURE ALL TASKS HAVE RELEVANT CONTEXT ABOUT THE CODE TO BE WRITTEN, INCLUDE DETAILS ON HOW TO CALL FUNCTIONS, CLASSES, IMPORTS, ETC. AGENTS HAVE NO VIEW OF OTHER TASKS, SO THEY NEED TO BE SELF-CONTAINED. RETURN THE JSON:`,
								"outputJSON": true,
							},
						},
					},
					{
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"template": `You are an AGI agent responsible for improving a list of tasks in JSON format and adding ALL the necessary details to each task. These tasks will be executed individually by agents that have no idea about other tasks or what code exists in the codebase. It is FUNDAMENTAL that each task has enough details so that an individual isolated agent can execute. The metadata of the task is the only information the agents will have.

    Each task should contain the details necessary to execute it. For example, if it creates a function, it needs to contain the details about the arguments to be used in that function and this needs to be consistent across all tasks.

    Look at all tasks at once, and update the task description adding details to it for each task so that it can be executed by an agent without seeing the other tasks and to ensure consistency across all tasks. DETAILS ARE CRUCIAL. For example, if one task creates a class, it should have all the details about the class, including the arguments to be used in the constructor. If another task creates a function that uses the class, it should have the details about the class and the arguments to be used in the constructor.

    RETURN JSON OUTPUTS ONLY.
    
    Here is the overall objective you need to refactor the tasks for: {{.Memory.Objective}}.
    Here is the task list you need to improve: {{.Text}}
    
    RETURN THE SAME TASK LIST but with the description improved to contain the details you is adding for each task in the list. DO NOT MAKE OTHER MODIFICATIONS TO THE LIST. Your input should go in the 'description' field of each task.
    
    RETURN JSON ONLY:`,
								"outputJSON": true,
							},
						},
					},
					{
						Executor: &anyi.ExecutorConfig{
							Type: "llm",
							WithConfig: map[string]interface{}{
								"outputJSON": true,
								"template": `You are an AGI agent responsible for improving a list of tasks in JSON format and adding ALL the necessary context to each task's description property. These tasks will be executed individually by agents that have no idea about other tasks or what code exists in the codebase. It is FUNDAMENTAL that each task has enough context so that an individual isolated agent can execute. The metadata of the task is the only information the agents will have.

    Look at all tasks at once, and add the necessary context to each task so that it can be executed by an agent without seeing the other tasks. Remember, one agent can only see one task and has no idea about what happened in other tasks. CONTEXT IS CRUCIAL. For example, if one task creates one folder and the other tasks creates a file in that folder. The second tasks should contain the name of the folder that already exists and the information that it already exists.

    This is even more important for tasks that require importing functions, classes, etc. If a task needs to call a function or initialize a Class, it needs to have the detailed arguments, etc.

    Note that you should identify when imports need to happen and specify this in the context. Also, you should identify when functions/classes/etc already exist and specify this very clearly because the agents sometimes duplicate things not knowing.

    Always use imports with the file name. For example, 'from my_script import MyScript'. 
    
    RETURN JSON OUTPUTS ONLY.

	ONLY UPDATE THE DESCRIPTION FIELD OF EACH TASK. DO NOT MAKE OTHER MODIFICATIONS TO THE TASK LIST.
    
    Here is the overall objective you need to refactor the tasks for: {{.Memory.Objective}}.
    Here is the task list you need to improve: {{.Text}}
    
    RETURN THE SAME TASK LIST but with a new field called 'isolated_context' for each task in the list. This field should be a string with the context you are adding. DO NOT MAKE OTHER MODIFICATIONS TO THE LIST.
    
    RETURN JSON ONLY:`,
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

type InitialTasksPlan struct {
	Tasks []TaskData `json:"tasks"`
}

func Example_coderTaskAGI() {

	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		PadLevelText:  true,
		FullTimestamp: false,
		DisableQuote:  true,
	})
	task := Task{
		Objective: `Create a Python program that consists of a single class named 'TemperatureConverter' in a file named 'temperature_converter.py'. The class should have the following methods:
- celsius_to_fahrenheit(self, celsius: float) -> float: Converts Celsius temperature to Fahrenheit.
- fahrenheit_to_celsius(self, fahrenheit: float) -> float: Converts Fahrenheit temperature to Celsius.
Create a separate 'main.py' file that imports the 'TemperatureConverter' class, takes user input for the temperature value and the unit, converts the temperature to the other unit, and then prints the result.`,
	}

	InitAnyi()
	loop := true

	initialFlow, err := anyi.GetFlow("taskInitiator")
	if err != nil {
		log.Fatal(err)
	}
	context := &flow.FlowContext{
		Memory: task,
	}
	context, err = initialFlow.Run(*context)
	if err != nil {
		log.Fatal(err)
		panic("error in running flow")
	}

	taskPlan := &InitialTasksPlan{}
	err = context.UnmarshalJsonText(&taskPlan)

	if err != nil {
		log.Fatal(err)
		panic("error in unmarshalling")
	}

	for loop == true {
		/**
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
		*/
	}
}
