package coder

import (
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
)

type TaskData struct {
	Id              int    `json:"id"`
	Description     string `json:"description"`
	FilePath        string `json:"file_path"`
	IsolatedContext string `json:"isolated_context"`
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

type TaskPlan struct {
	Tasks       []TaskData `json:"tasks"`
	Objective   string     `json:"objective"`
	CurrentTask *TaskData  `json:"currentTask"`
	OS          string     `json:"os"`
}

func InitAnyi() {
	anyi.ConfigFromFile("config.toml")
}

const REPOSITORY = "playground"

func Example_coderTaskAGI() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		PadLevelText:  true,
		FullTimestamp: false,
		DisableQuote:  true,
	})

	plan := TaskPlan{
		Objective: `Create a Python program that consists of a single class named 'TemperatureConverter' in a file named 'temperature_converter.py'. The class should have the following methods:
- celsius_to_fahrenheit(self, celsius: float) -> float: Converts Celsius temperature to Fahrenheit.
- fahrenheit_to_celsius(self, fahrenheit: float) -> float: Converts Fahrenheit temperature to Celsius.
Create a separate 'main.py' file that imports the 'TemperatureConverter' class, takes user input for the temperature value and the unit, converts the temperature to the other unit, and then prints the result.`,
		OS: runtime.GOOS,
	}

	log.Info("****Objective****")
	log.Info(plan.Objective)

	InitAnyi()

	// Ensure playground directory exists
	_, err := os.Stat(REPOSITORY)
	if os.IsNotExist(err) {
		err = os.Mkdir(REPOSITORY, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Phase 1: Task generation and optimization
	log.Info("*****Starting Task Processing*****")
	log.Info(" - Creating initial tasks")

	// 1. Create initial task list
	initialFlow, err := anyi.GetFlow("taskInitFlow")
	if err != nil {
		log.Fatal("Failed to get taskInitFlow: ", err)
	}

	context := &flow.FlowContext{
		Memory: plan,
	}

	context, err = initialFlow.Run(*context)
	if err != nil {
		log.Fatal("Failed to run taskInitFlow: ", err)
	}

	err = context.UnmarshalJsonText(&plan)
	if err != nil {
		log.Fatal("Failed to parse task list: ", err)
	}

	log.Info(" - Task creation completed, total", len(plan.Tasks), "tasks")
	log.Info("*****Task List*****")
	for i, task := range plan.Tasks {
		log.Infof("Task %d: %s", i+1, task.Description)
		if task.IsolatedContext != "" {
			log.Infof("  Context: %s", task.IsolatedContext)
		}
		log.Infof("  File Path: %s", task.FilePath)
	}

	// Phase 2: Execute each task
	log.Info("*****Starting Task Execution*****")
	for i, task := range plan.Tasks {
		log.Infof("Executing task %d/%d: %s", i+1, len(plan.Tasks), task.Description)
		plan.CurrentTask = &task

		executeFlow, err := anyi.GetFlow("taskExecuteFlow")
		if err != nil {
			log.Errorf("Failed to get taskExecuteFlow: %v", err)
			continue
		}

		context := &flow.FlowContext{
			Memory: &plan,
		}

		context, err = executeFlow.Run(*context)
		if err != nil {
			log.Errorf("Failed to execute task: %v", err)
			continue
		}

		log.Infof("Task executed successfully! Result: %s", context.Text)
		log.Info("---")
	}

	log.Info("*****All Tasks Completed*****")

	// Show final results
	log.Info("*****Final Results*****")
	files, err := os.ReadDir(REPOSITORY)
	if err != nil {
		log.Error("Failed to read playground directory: ", err)
	} else {
		log.Info("Generated files:")
		for _, file := range files {
			log.Infof("  - %s", file.Name())
		}
	}
}

// Add a simplified version for quick testing
func Example_simpleCoderTaskAGI() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		PadLevelText:  true,
		FullTimestamp: false,
		DisableQuote:  true,
	})

	plan := TaskPlan{
		Objective: `Create a simple Python hello world program in a file named 'hello.py' that prints "Hello, World!" to the console.`,
		OS:        runtime.GOOS,
	}

	InitAnyi()

	// Ensure playground directory exists
	_, err := os.Stat(REPOSITORY)
	if os.IsNotExist(err) {
		err = os.Mkdir(REPOSITORY, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Info("Objective: ", plan.Objective)

	// Generate tasks
	initialFlow, err := anyi.GetFlow("taskInitFlow")
	if err != nil {
		log.Fatal("Failed to get taskInitFlow: ", err)
	}

	context := &flow.FlowContext{
		Memory: plan,
	}

	context, err = initialFlow.Run(*context)
	if err != nil {
		log.Fatal("Failed to run taskInitFlow: ", err)
	}

	err = context.UnmarshalJsonText(&plan)
	if err != nil {
		log.Fatal("Failed to parse task list: ", err)
	}

	// Execute tasks
	for _, task := range plan.Tasks {
		log.Info("Executing task: ", task.Description)
		plan.CurrentTask = &task

		executeFlow, err := anyi.GetFlow("taskExecuteFlow")
		if err != nil {
			log.Error("Failed to get taskExecuteFlow: ", err)
			continue
		}

		context := &flow.FlowContext{
			Memory: &plan,
		}

		context, err = executeFlow.Run(*context)
		if err != nil {
			log.Error("Failed to execute task: ", err)
			continue
		}

		log.Info("Task completed")
	}

	log.Info("All tasks completed")
}

// Manual task definition example for testing specific functionality
func Example_manualTaskAGI() {
	log.SetLevel(log.InfoLevel)

	InitAnyi()

	// Ensure playground directory exists
	_, err := os.Stat(REPOSITORY)
	if os.IsNotExist(err) {
		err = os.Mkdir(REPOSITORY, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Manually create a simple task plan
	plan := TaskPlan{
		Objective: "Create a simple calculator",
		OS:        runtime.GOOS,
		Tasks: []TaskData{
			{
				Id:              1,
				Description:     "Write code to create a Python calculator with add, subtract, multiply, and divide functions",
				FilePath:        "calculator.py",
				IsolatedContext: "Create a simple calculator class with basic arithmetic operations. The class should be named 'Calculator' and have methods: add(a, b), subtract(a, b), multiply(a, b), divide(a, b). Each method should return the result of the operation.",
			},
		},
	}

	log.Info("Manual task plan:")
	for _, task := range plan.Tasks {
		log.Info("Task: ", task.Description)
		plan.CurrentTask = &task

		executeFlow, err := anyi.GetFlow("taskExecuteFlow")
		if err != nil {
			log.Error("Failed to get taskExecuteFlow: ", err)
			continue
		}

		context := &flow.FlowContext{
			Memory: &plan,
		}

		context, err = executeFlow.Run(*context)
		if err != nil {
			log.Error("Failed to execute task: ", err)
			continue
		}

		log.Info("Task completed, result: ", context.Text)
	}
}
