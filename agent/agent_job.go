package agent

import (
	"encoding/json"
	"errors"

	"github.com/jieliu2000/anyi/agent/agentmodel"
	"github.com/jieliu2000/anyi/agentflows/model"
	"github.com/jieliu2000/anyi/registry"
	log "github.com/sirupsen/logrus"
)

// PlanningStepResult represents a single step in the planning result
type PlanningStepResult struct {
	FlowName    string `json:"flowName"`
	Description string `json:"description"`
}

// StartJob starts a new job for the agent with the given context
// It returns an AgentJob reference immediately while the job runs asynchronously
func StartAgentJob(a *agentmodel.Agent, context *agentmodel.AgentContext) (*agentmodel.AgentJob, error) {
	// Check if agent has at least one flow
	if len(a.Flows) == 0 {
		return nil, errors.New("agent must have at least one flow to start a job")
	}

	if a.Client == nil {
		return nil, errors.New("agent must have a valid client to start a job")
	}

	job := &agentmodel.AgentJob{
		Agent:   a,
		Context: context,
		Status:  "running",
	}

	// Run the job asynchronously
	go ExecuteJob(job)

	return job, nil
}

// Execute runs the agent job asynchronously
func ExecuteJob(job *agentmodel.AgentJob) {
	// Initialize stop channel if not already done

	var taskPlan = PlanJobTasks(job)

	for _, task := range taskPlan {

		// Execute each task in the plan
		RunJobTask(job, task)
	}

	job.Status = "completed"
}

func RunJobTask(job *agentmodel.AgentJob, task string) error {
	// task parameter is a JSON formatted planning step that needs to be parsed
	var stepResult PlanningStepResult
	_, err := registry.GetFlow("task")

	if err != nil {
		log.Errorf("Failed to get flow: %v", err)
		return err
	}
	err = json.Unmarshal([]byte(task), &stepResult)
	if err != nil {
		log.Errorf("Failed to unmarshal task: %v", err)
		return err
	}

	// Find the corresponding flow based on stepResult.FlowName and execute it
	// Here we only log for now, actual execution logic can be added later
	log.Infof("Running task: %s - %s", stepResult.FlowName, stepResult.Description)
	return nil
}

func PlanJobTasks(job *agentmodel.AgentJob) []string {
	// Convert data from job and its associated Agent to AgentPlanningData structure
	planningData := model.AgentPlanningData{
		Role:              job.Agent.Role,
		BackStory:         job.Agent.BackStory,
		PreferredLanguage: job.Agent.PreferredLanguage,
		Goal:              job.Context.Goal,
	}

	// Convert AvailableFlows
	planningData.AvailableFlows = make([]model.FlowInfo, len(job.Agent.Flows))
	for i, flow := range job.Agent.Flows {
		planningData.AvailableFlows[i] = model.FlowInfo{
			Name:        flow.Name,
			Description: flow.Description,
		}
	}

	// Call RunPlanningFlow
	planningText, err := RunPlanningFlow(planningData, job.Agent.Client)
	if err != nil {
		log.Errorf("Failed to run planning flow: %v", err)
		return []string{}
	}

	// Parse planning results
	var planningResults []PlanningStepResult
	err = json.Unmarshal([]byte(planningText), &planningResults)
	if err != nil {
		log.Errorf("Failed to unmarshal planning results: %v", err)
		return []string{}
	}

	// Convert planning results to string array and return
	taskPlan := make([]string, len(planningResults))
	for i, result := range planningResults {
		taskBytes, _ := json.Marshal(result)
		taskPlan[i] = string(taskBytes)
	}

	return taskPlan
}
