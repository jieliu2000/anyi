package agi

import "log"

type TaskData struct {
	Description string
}

func Example_taskAGI() {

	objective := "Use python to create an AI digital employee project which can generate code for Quasar hybrid mobile app based on user input requirements."

	loop := true
	tasks := []TaskData{{
		Description: "Create a new project in python",
	}}

	for loop == true {
		if len(tasks) > 0 {

		} else {
			log.Println("All tasks completed!")
			loop = false
		}

	}
}
