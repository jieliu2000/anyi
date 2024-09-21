package coder

import (
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
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
	}

	InitAnyi()
	initialFlow, err := anyi.GetFlow("taskInitFlow")
	if err != nil {
		log.Fatal(err)
	}
	context := &flow.FlowContext{
		Memory: plan,
	}
	context, err = initialFlow.Run(*context)
	if err != nil {
		log.Fatal(err)
		panic("error in running flow")
	}

	err = context.UnmarshalJsonText(&plan)

	if err != nil {
		log.Fatal(err)
		panic("error in unmarshalling")
	}
	plan.OS = runtime.GOOS

	_, err = os.Stat(REPOSITORY)

	if os.IsNotExist(err) {
		err = os.Mkdir(REPOSITORY, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, task := range plan.Tasks {
		log.Infof("Task: %s", task.Description)
		plan.CurrentTask = &task

		executeFlow, _ := anyi.GetFlow("taskExecuteFlow")
		context := &flow.FlowContext{
			Memory: &plan,
		}
		context, err = executeFlow.Run(*context)

		log.Info("Executed task successfully!, result:", context.Text)
	}
}
