package agent

import (
	"testing"
	"time"

	"github.com/jieliu2000/anyi/agent/agentmodel"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestAgentJob_Execute(t *testing.T) {
	// Create a mock agent
	agent := &agentmodel.Agent{
		Role: "test-agent",
	}

	// Create a mock context
	context := &agentmodel.AgentContext{}

	// Create a job
	job := &agentmodel.AgentJob{
		Agent:   agent,
		Context: context,
		Status:  "pending",
	}

	// Execute the job
	ExecuteJob(job)

	// Check that the job completed
	assert.Equal(t, "completed", job.Status)
}

func TestAgentJob_Stop(t *testing.T) {
	// Create a mock agent
	agent := &agentmodel.Agent{
		Role: "test-agent",
	}

	// Create a mock context
	context := &agentmodel.AgentContext{}

	// Create a job
	job := &agentmodel.AgentJob{
		Agent:   agent,
		Context: context,
		Status:  "running",
	}

	// Stop the job
	err := StopJob(job)
	assert.NoError(t, err)

	// Check that the job is paused
	assert.Equal(t, "paused", job.Status)
}

func TestAgentJob_Resume(t *testing.T) {
	// Create a mock agent
	agent := &agentmodel.Agent{
		Role: "test-agent",
	}

	// Create a mock context
	context := &agentmodel.AgentContext{}

	// Create a job
	job := &agentmodel.AgentJob{
		Agent:   agent,
		Context: context,
		Status:  "paused",
	}

	// Resume the job
	err := ResumeJob(job)
	assert.NoError(t, err)

	// Check that the job is running
	assert.Equal(t, "running", job.Status)

	// Give some time for the goroutine to start
	time.Sleep(10 * time.Millisecond)
}

func TestAgentJob_StopDuringExecution(t *testing.T) {
	// Create a mock agent
	agent := &agentmodel.Agent{
		Role: "test-agent",
	}

	// Create a mock context
	context := &agentmodel.AgentContext{}

	// Create a job with a long-running task plan
	job := &agentmodel.AgentJob{
		Agent:   agent,
		Context: context,
		Status:  "running",
	}

	// Start execution in a goroutine
	go func() {
		ExecuteJob(job)
	}()

	// Give some time for execution to start
	time.Sleep(5 * time.Millisecond)

	// Stop the job
	err := StopJob(job)
	assert.NoError(t, err)

	// Check that the job is paused
	assert.Equal(t, "paused", job.Status)
}

func TestAgentJob_PlanTasks(t *testing.T) {
	// Create a mock agent
	agent := &agentmodel.Agent{
		Role: "test-agent",
	}

	// Create a mock context
	context := &agentmodel.AgentContext{}

	// Create a job
	job := &agentmodel.AgentJob{
		Agent:   agent,
		Context: context,
		Status:  "pending",
	}

	// Plan tasks
	tasks := PlanJobTasks(job)

	// Currently PlanTasks returns an empty slice, so we check for that
	assert.Empty(t, tasks)
}

func TestAgentJob_RunTask(t *testing.T) {
	// Create a mock agent
	agent := &agentmodel.Agent{
		Role: "test-agent",
	}

	// Create a mock context
	context := &agentmodel.AgentContext{}

	// Create a job
	job := &agentmodel.AgentJob{
		Agent:   agent,
		Context: context,
		Status:  "running",
	}

	// Run a task (this is a placeholder implementation)
	RunJobTask(job, "test-task")

	// Currently RunTask does nothing, so we just verify it doesn't panic
	assert.True(t, true)
}

func TestAgent_StartJob_CompleteScenario(t *testing.T) {
	// Create an agent without client - should return error
	agentWithoutClient := &agentmodel.Agent{
		Role: "test-agent",
		// Client is nil by default
		Flows: []*flow.Flow{
			{
				Name: "test-flow",
			},
		},
	}

	// Create a context with a goal
	context := &agentmodel.AgentContext{
		Goal:            "Test complete scenario",
		ShortTermMemory: make(map[string]interface{}),
		ExecuteLog:      []string{},
	}

	// Try to start a job without client - should return error
	job, err := StartAgentJob(agentWithoutClient, context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	// Try to start a job without flows - should return error
	job, err = StartAgentJob(agentWithoutClient, context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	// Create an agent with client but without flows - should return error
	agentWithClientNoFlows := &agentmodel.Agent{
		Role:   "test-agent",
		Client: &test.MockClient{},
	}

	// Try to start a job with client but without flows - should return error
	job, err = StartAgentJob(agentWithClientNoFlows, context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	// Create a mock flow
	client := &test.MockClient{}
	mockFlow, err := flow.NewFlow(client, "test-flow")
	assert.NoError(t, err)

	// Create an agent with both client and flows - should succeed
	agent := &agentmodel.Agent{
		Role:   "test-agent",
		Client: &test.MockClient{},
		Flows:  []*flow.Flow{mockFlow},
	}

	// Start a job using the agent's StartJob method
	job, err = StartAgentJob(agent, context)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, agent, job.Agent)
	assert.Equal(t, context, job.Context)
	assert.Equal(t, "running", job.Status)

	// Give some time for the job to start executing
	time.Sleep(5 * time.Millisecond)

	// Check job status (should be completed since PlanTasks returns empty slice)
	assert.Equal(t, "completed", job.Status)
}

func TestAgent_StartJob_StopScenario(t *testing.T) {
	// Create an agent without client - should return error
	agentWithoutClient := &agentmodel.Agent{
		Role: "test-agent",
		// Client is nil by default
		Flows: []*flow.Flow{
			{
				Name: "test-flow",
			},
		},
	}

	// Create a context
	context := &agentmodel.AgentContext{
		Goal:            "Test stop scenario",
		ShortTermMemory: make(map[string]interface{}),
		ExecuteLog:      []string{},
	}

	// Try to start a job without client - should return error
	job, err := StartAgentJob(agentWithoutClient, context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	// Create an agent with client but without flows - should return error
	agentWithClientNoFlows := &agentmodel.Agent{
		Role:   "test-agent",
		Client: &test.MockClient{},
	}

	// Try to start a job with client but without flows - should return error
	job, err = StartAgentJob(agentWithClientNoFlows, context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	// Create a mock flow
	client := &test.MockClient{}
	mockFlow, err := flow.NewFlow(client, "test-flow")
	assert.NoError(t, err)

	// Create an agent with both client and flows
	agent := &agentmodel.Agent{
		Role:   "test-agent",
		Client: &test.MockClient{},
		Flows:  []*flow.Flow{mockFlow},
	}

	// Start a job
	job, err = StartAgentJob(agent, context)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "running", job.Status)

	// Give some time for the job to start
	time.Sleep(1 * time.Millisecond)

	// Stop the job
	err = StopJob(job)
	assert.NoError(t, err)
	assert.Equal(t, "paused", job.Status)

	// Try to resume the job
	err = ResumeJob(job)
	assert.NoError(t, err)
	assert.Equal(t, "running", job.Status)

	// Give some time for the resumed job to run
	time.Sleep(5 * time.Millisecond)

	// Job should complete quickly since PlanTasks returns empty slice
	assert.Equal(t, "completed", job.Status)
}

func TestAgent_StartJob_ImmediateStop(t *testing.T) {
	// Create an agent without client - should return error
	agentWithoutClient := &agentmodel.Agent{
		Role: "test-agent",
		// Client is nil by default
		Flows: []*flow.Flow{
			{
				Name: "test-flow",
			},
		},
	}

	// Create a context
	context := &agentmodel.AgentContext{
		Goal:            "Test immediate stop",
		ShortTermMemory: make(map[string]interface{}),
		ExecuteLog:      []string{},
	}

	// Try to start a job without client - should return error
	_, err := StartAgentJob(agentWithoutClient, context)
	assert.Error(t, err)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	// Create an agent with client but without flows - should return error
	agentWithClientNoFlows := &agentmodel.Agent{
		Role:   "test-agent",
		Client: &test.MockClient{},
	}

	// Try to start a job with client but without flows - should return error
	_, err = StartAgentJob(agentWithClientNoFlows, context)
	assert.Error(t, err)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	// Create a mock flow
	client := &test.MockClient{}
	mockFlow, err := flow.NewFlow(client, "test-flow")
	assert.NoError(t, err)

	// Create an agent with both client and flows
	agent := &agentmodel.Agent{
		Role:   "test-agent",
		Client: &test.MockClient{},
		Flows:  []*flow.Flow{mockFlow},
	}

	// Start a job
	job, err := StartAgentJob(agent, context)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "running", job.Status)

	// Immediately stop the job
	err = StopJob(job)
	assert.NoError(t, err)
	assert.Equal(t, "paused", job.Status)

	// Verify that stop channel is properly closed by trying to resume
	err = ResumeJob(job)
	assert.NoError(t, err)
	assert.Equal(t, "running", job.Status)

	// Give some time to complete
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, "completed", job.Status)
}

func TestAgent_StartJob(t *testing.T) {
	// Test case 1: Create an agent without flows - should return error
	agent := &agentmodel.Agent{
		Role:              "Test Agent",
		Client:            &test.MockClient{},
		PreferredLanguage: "English",
		BackStory:         "A test agent",
		Flows:             []*flow.Flow{},
	}

	// Create a context
	context := &agentmodel.AgentContext{
		Goal:            "Test the StartJob function",
		ShortTermMemory: make(map[string]interface{}),
		ExecuteLog:      []string{},
	}

	// Try to start a job without flows - should return error
	job, err := StartAgentJob(agent, context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	// Test case 2: Create an agent without client - should return error
	agentWithoutClient := &agentmodel.Agent{
		Role:              "Test Agent",
		Client:            nil,
		PreferredLanguage: "English",
		BackStory:         "A test agent",
		Flows:             []*flow.Flow{},
	}

	// Try to start a job without client - should return error
	job, err = StartAgentJob(agentWithoutClient, context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	// Test case 3: Create an agent with flows but without client - should return error
	client := &test.MockClient{}
	mockFlow, err := flow.NewFlow(client, "test-flow")
	assert.NoError(t, err)

	agentWithFlowNoClient := &agentmodel.Agent{
		Role:              "Test Agent",
		Client:            nil,
		PreferredLanguage: "English",
		BackStory:         "A test agent",
		Flows:             []*flow.Flow{mockFlow},
	}

	// Try to start a job with flows but without client - should return error
	job, err = StartAgentJob(agentWithFlowNoClient, context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	// Test case 4: Create an agent with flows and client - should succeed
	agentWithFlowAndClient := &agentmodel.Agent{
		Role:              "Test Agent",
		Client:            &test.MockClient{},
		PreferredLanguage: "English",
		BackStory:         "A test agent",
		Flows:             []*flow.Flow{mockFlow},
	}

	// Start a job with flows and client - should succeed
	job, err = StartAgentJob(agentWithFlowAndClient, context)
	assert.NoError(t, err)
	assert.NotNil(t, job)

	// Give some time for the goroutine to execute
	time.Sleep(10 * time.Millisecond)

	// Check that we got a job back
	assert.NotNil(t, job)
	assert.Equal(t, agentWithFlowAndClient, job.Agent)
	assert.Equal(t, context, job.Context)

	// Job should be completed since PlanTasks returns an empty slice
	assert.Equal(t, "completed", job.Status)
}
func TestAgent_MultipleJobs(t *testing.T) {
	// Create an agent without client - should return error
	agentWithoutClient := &agentmodel.Agent{
		Role: "test-agent",
		// Client is nil by default
		Flows: []*flow.Flow{
			{
				Name:  "test-flow",
				Steps: []flow.Step{},
			},
		},
	}

	// Create multiple contexts
	context1 := &agentmodel.AgentContext{
		Goal:            "Test job 1",
		ShortTermMemory: make(map[string]interface{}),
		ExecuteLog:      []string{},
	}

	context2 := &agentmodel.AgentContext{
		Goal:            "Test job 2",
		ShortTermMemory: make(map[string]interface{}),
		ExecuteLog:      []string{},
	}

	// Try to start jobs without client - should return error
	job1, err := StartAgentJob(agentWithoutClient, context1)
	assert.Error(t, err)
	assert.Nil(t, job1)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	job2, err := StartAgentJob(agentWithoutClient, context2)
	assert.Error(t, err)
	assert.Nil(t, job2)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	// Create an agent with client but without flows - should return error
	agentWithClientNoFlows := &agentmodel.Agent{
		Role:   "test-agent",
		Client: &test.MockClient{},
	}

	// Try to start jobs with client but without flows - should return error
	job1, err = StartAgentJob(agentWithClientNoFlows, context1)
	assert.Error(t, err)
	assert.Nil(t, job1)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	job2, err = StartAgentJob(agentWithClientNoFlows, context2)
	assert.Error(t, err)
	assert.Nil(t, job2)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	// Create a mock flow
	client := &test.MockClient{}
	mockFlow, err := flow.NewFlow(client, "test-flow")
	assert.NoError(t, err)

	// Create an agent with both client and flows
	agent := &agentmodel.Agent{
		Role:   "test-agent",
		Client: &test.MockClient{},
		Flows:  []*flow.Flow{mockFlow},
	}

	// Start multiple jobs
	job1, err = StartAgentJob(agent, context1)
	assert.NoError(t, err)
	assert.NotNil(t, job1)

	job2, err = StartAgentJob(agent, context2)
	assert.NoError(t, err)
	assert.NotNil(t, job2)

	assert.NotEqual(t, job1, job2)

	// Both jobs should be running
	//assert.Equal(t, "running", job1.Status)
	//assert.Equal(t, "running", job2.Status)

	// Give some time for jobs to complete
	time.Sleep(10 * time.Millisecond)

	// Both jobs should be completed
	assert.Equal(t, "completed", job1.Status)
	assert.Equal(t, "completed", job2.Status)
}
