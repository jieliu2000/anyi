package main

import (
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/jieliu2000/anyi"
	"github.com/jieliu2000/anyi/flow"
)

type TaskPlan struct {
	Tasks        []TaskData `json:"tasks"`
	Objective    string     `json:"objective"`
	CurrentTask  *TaskData  `json:"currentTask"`
	OS           string     `json:"os"`
	Instructions []string   `json:"instructions"`
}

type TaskData struct {
	Id          int    `json:"id"`
	Description string `json:"description"`

	FilePath string `json:"file_path"`

	IsolatedContext string `json:"isolated_context"`
}

func InitAnyi() {

	anyi.ConfigFromFile("config.toml")

}

var REPOSITORY = "docs"

func main() {

	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{
		PadLevelText:  true,
		FullTimestamp: false,
		DisableQuote:  true,
	})
	documentationPlan := TaskPlan{
		Objective: `Create an comphensive documentation for the open source golang project. The documentation could be several linked files in markdown format.`,
		Instructions: []string{
			"Go through the project and plan how to write the documentation first",
			"The documentation's audience is for the developers who have some basic knowledge of golang and LLM.",
			"The documentation should be easy to read and understand.",
			"There should be breif introductions about the project. There should be other parts about the detailed APIs and concepts of the project in the documentation.",
		},
	}

	InitAnyi()
	initialFlow, err := anyi.GetFlow("taskInitFlow")
	if err != nil {
		log.Fatal(err)
	}
	context := &flow.FlowContext{
		Text:   "Go through the project and plan how to write the documentation, including ",
		Memory: documentationPlan,
	}
	context, err = initialFlow.Run(*context)
	if err != nil {
		log.Fatal(err)
		panic("error in running flow")
	}

	err = context.UnmarshalJsonText(&documentationPlan)

	if err != nil {
		log.Fatal(err)
		panic("error in unmarshalling")
	}
	documentationPlan.OS = runtime.GOOS

	_, err = os.Stat(REPOSITORY)

	if os.IsNotExist(err) {
		err = os.Mkdir(REPOSITORY, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, task := range documentationPlan.Tasks {
		log.Infof("Task: %s", task.Description)
		documentationPlan.CurrentTask = &task

		executeFlow, _ := anyi.GetFlow("taskExecuteFlow")
		context := &flow.FlowContext{
			Memory: &documentationPlan,
		}
		context, err = executeFlow.Run(*context)
		if err != nil {
			log.Fatal(err)
		}

		log.Info("Executed task successfully!, result:", context.Text)
	}

}
