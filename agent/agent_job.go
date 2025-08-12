package agent

import (
	"encoding/json"
	"errors"

	"github.com/jieliu2000/anyi/agent/agentmodel"
	"github.com/jieliu2000/anyi/agentflows/model"
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

func RunJobTask(job *agentmodel.AgentJob, task string) {
	// task参数是JSON格式的规划步骤，需要解析它
	var stepResult PlanningStepResult
	err := json.Unmarshal([]byte(task), &stepResult)
	if err != nil {
		log.Errorf("Failed to unmarshal task: %v", err)
		return
	}

	// 根据stepResult.FlowName查找对应的flow并执行
	// 这里暂时只记录日志，实际执行逻辑可以后续补充
	log.Infof("Running task: %s - %s", stepResult.FlowName, stepResult.Description)
}

func PlanJobTasks(job *agentmodel.AgentJob) []string {
	// 将job以及job相关Agent中的数据转化为AgentPlanningData结构
	planningData := model.AgentPlanningData{
		Role:              job.Agent.Role,
		BackStory:         job.Agent.BackStory,
		PreferredLanguage: job.Agent.PreferredLanguage,
		Goal:              job.Context.Goal,
	}

	// 转换AvailableFlows
	planningData.AvailableFlows = make([]model.FlowInfo, len(job.Agent.Flows))
	for i, flow := range job.Agent.Flows {
		planningData.AvailableFlows[i] = model.FlowInfo{
			Name:        flow.Name,
			Description: flow.Description,
		}
	}

	// 调用RunPlanningFlow
	planningText, err := RunPlanningFlow(planningData)
	if err != nil {
		log.Errorf("Failed to run planning flow: %v", err)
		return []string{}
	}

	// 解析规划结果
	var planningResults []PlanningStepResult
	err = json.Unmarshal([]byte(planningText), &planningResults)
	if err != nil {
		log.Errorf("Failed to unmarshal planning results: %v", err)
		return []string{}
	}

	// 将规划结果转换为字符串数组返回
	taskPlan := make([]string, len(planningResults))
	for i, result := range planningResults {
		taskBytes, _ := json.Marshal(result)
		taskPlan[i] = string(taskBytes)
	}

	return taskPlan
}

// Resume continues a paused job
func ResumeJob(job *agentmodel.AgentJob) error {
	// When resuming, we replan based on the existing context
	job.Status = "running"

	ExecuteJob(job)
	return nil
}

// Stop pauses the job execution
func StopJob(job *agentmodel.AgentJob) error {
	job.Status = "paused"

	return nil
}
