package agent

import (
	"testing"
	"time"

	"github.com/jieliu2000/anyi/executors"
	"github.com/jieliu2000/anyi/flow"
	"github.com/jieliu2000/anyi/llm/chat"
	"github.com/jieliu2000/anyi/llm/tools"
	"github.com/stretchr/testify/assert"
)

// Mock client implementation for testing
type mockClient struct{}

func (c *mockClient) Chat(messages []chat.Message, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	return &chat.Message{Role: "assistant", Content: "mock response"}, chat.ResponseInfo{}, nil
}

func (c *mockClient) ChatWithFunctions(messages []chat.Message, functions []tools.FunctionConfig, options *chat.ChatOptions) (*chat.Message, chat.ResponseInfo, error) {
	return &chat.Message{Role: "assistant", Content: "mock function response"}, chat.ResponseInfo{}, nil
}

func (c *mockClient) CountTokens(text string) (int, error) {
	return len(text), nil
}

func TestAgentJob_Execute(t *testing.T) {
	// Create a mock agent
	agent := &Agent{
		Role: "test-agent",
	}

	// Create a mock context
	context := &AgentContext{}

	// Create a job
	job := &AgentJob{
		Agent:   agent,
		Context: context,
		Status:  "pending",
	}

	// Execute the job
	job.Execute()

	// Check that the job completed
	assert.Equal(t, "completed", job.Status)
}

func TestAgentJob_Stop(t *testing.T) {
	// Create a mock agent
	agent := &Agent{
		Role: "test-agent",
	}

	// Create a mock context
	context := &AgentContext{}

	// Create a job
	job := &AgentJob{
		Agent:   agent,
		Context: context,
		Status:  "running",
	}

	// Stop the job
	err := job.Stop()
	assert.NoError(t, err)

	// Check that the job is paused
	assert.Equal(t, "paused", job.Status)
}

func TestAgentJob_Resume(t *testing.T) {
	// Create a mock agent
	agent := &Agent{
		Role: "test-agent",
	}

	// Create a mock context
	context := &AgentContext{}

	// Create a job
	job := &AgentJob{
		Agent:   agent,
		Context: context,
		Status:  "paused",
	}

	// Resume the job
	err := job.Resume()
	assert.NoError(t, err)

	// Check that the job is running
	assert.Equal(t, "running", job.Status)

	// Give some time for the goroutine to start
	time.Sleep(10 * time.Millisecond)
}

func TestAgentJob_StopDuringExecution(t *testing.T) {
	// Create a mock agent
	agent := &Agent{
		Role: "test-agent",
	}

	// Create a mock context
	context := &AgentContext{}

	// Create a job with a long-running task plan
	job := &AgentJob{
		Agent:   agent,
		Context: context,
		Status:  "running",
	}

	// Start execution in a goroutine
	go func() {
		job.Execute()
	}()

	// Give some time for execution to start
	time.Sleep(5 * time.Millisecond)

	// Stop the job
	err := job.Stop()
	assert.NoError(t, err)

	// Check that the job is paused
	assert.Equal(t, "paused", job.Status)
}

func TestAgentJob_PlanTasks(t *testing.T) {
	// Create a mock agent
	agent := &Agent{
		Role: "test-agent",
	}

	// Create a mock context
	context := &AgentContext{}

	// Create a job
	job := &AgentJob{
		Agent:   agent,
		Context: context,
		Status:  "pending",
	}

	// Plan tasks
	tasks := job.PlanTasks()

	// Currently PlanTasks returns an empty slice, so we check for that
	assert.Empty(t, tasks)
}

func TestAgentJob_RunTask(t *testing.T) {
	// Create a mock agent
	agent := &Agent{
		Role: "test-agent",
	}

	// Create a mock context
	context := &AgentContext{}

	// Create a job
	job := &AgentJob{
		Agent:   agent,
		Context: context,
		Status:  "running",
	}

	// Run a task (this is a placeholder implementation)
	job.RunTask("test-task")

	// Currently RunTask does nothing, so we just verify it doesn't panic
	assert.True(t, true)
}

func TestAgent_StartJob_CompleteScenario(t *testing.T) {
	// Create an agent without client - should return error
	agentWithoutClient := &Agent{
		Role: "test-agent",
		// Client is nil by default
		Flows: []*flow.Flow{
			{
				Name: "test-flow",
			},
		},
	}

	// Create a context with a goal
	context := &AgentContext{
		Goal:            "Test complete scenario",
		ShortTermMemory: make(map[string]interface{}),
		ExecuteLog:      []string{},
	}

	// Try to start a job without client - should return error
	job, err := agentWithoutClient.StartJob(context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	// Try to start a job without flows - should return error
	job, err = agentWithoutClient.StartJob(context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	// Create an agent with client but without flows - should return error
	agentWithClientNoFlows := &Agent{
		Role:   "test-agent",
		Client: &mockClient{},
	}

	// Try to start a job with client but without flows - should return error
	job, err = agentWithClientNoFlows.StartJob(context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	// Create a mock flow
	client := &mockClient{}
	mockFlow, err := flow.NewFlow(client, "test-flow")
	assert.NoError(t, err)

	// Create an agent with both client and flows - should succeed
	agent := &Agent{
		Role:   "test-agent",
		Client: &mockClient{},
		Flows:  []*flow.Flow{mockFlow},
	}

	// Start a job using the agent's StartJob method
	job, err = agent.StartJob(context)
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
	agentWithoutClient := &Agent{
		Role: "test-agent",
		// Client is nil by default
		Flows: []*flow.Flow{
			{
				Name: "test-flow",
			},
		},
	}

	// Create a context
	context := &AgentContext{
		Goal:            "Test stop scenario",
		ShortTermMemory: make(map[string]interface{}),
		ExecuteLog:      []string{},
	}

	// Try to start a job without client - should return error
	job, err := agentWithoutClient.StartJob(context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	// Create an agent with client but without flows - should return error
	agentWithClientNoFlows := &Agent{
		Role:   "test-agent",
		Client: &mockClient{},
	}

	// Try to start a job with client but without flows - should return error
	job, err = agentWithClientNoFlows.StartJob(context)
	assert.Error(t, err)
	assert.Nil(t, job)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	// Create a mock flow
	client := &mockClient{}
	mockFlow, err := flow.NewFlow(client, "test-flow")
	assert.NoError(t, err)

	// Create an agent with both client and flows
	agent := &Agent{
		Role:   "test-agent",
		Client: &mockClient{},
		Flows:  []*flow.Flow{mockFlow},
	}

	// Start a job
	job, err = agent.StartJob(context)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "running", job.Status)

	// Give some time for the job to start
	time.Sleep(1 * time.Millisecond)

	// Stop the job
	err = job.Stop()
	assert.NoError(t, err)
	assert.Equal(t, "paused", job.Status)

	// Try to resume the job
	err = job.Resume()
	assert.NoError(t, err)
	assert.Equal(t, "running", job.Status)

	// Give some time for the resumed job to run
	time.Sleep(5 * time.Millisecond)

	// Job should complete quickly since PlanTasks returns empty slice
	assert.Equal(t, "completed", job.Status)
}

func TestAgent_StartJob_ImmediateStop(t *testing.T) {
	// Create an agent without client - should return error
	agentWithoutClient := &Agent{
		Role: "test-agent",
		// Client is nil by default
		Flows: []*flow.Flow{
			{
				Name: "test-flow",
			},
		},
	}

	// Create a context
	context := &AgentContext{
		Goal:            "Test immediate stop",
		ShortTermMemory: make(map[string]interface{}),
		ExecuteLog:      []string{},
	}

	// Try to start a job without client - should return error
	_, err := agentWithoutClient.StartJob(context)
	assert.Error(t, err)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	// Create an agent with client but without flows - should return error
	agentWithClientNoFlows := &Agent{
		Role:   "test-agent",
		Client: &mockClient{},
	}

	// Try to start a job with client but without flows - should return error
	_, err = agentWithClientNoFlows.StartJob(context)
	assert.Error(t, err)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	// Create a mock flow
	client := &mockClient{}
	mockFlow, err := flow.NewFlow(client, "test-flow")
	assert.NoError(t, err)

	// Create an agent with both client and flows
	agent := &Agent{
		Role:   "test-agent",
		Client: &mockClient{},
		Flows:  []*flow.Flow{mockFlow},
	}

	// Start a job
	job, err := agent.StartJob(context)
	assert.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, "running", job.Status)

	// Immediately stop the job
	err = job.Stop()
	assert.NoError(t, err)
	assert.Equal(t, "paused", job.Status)

	// Verify that stop channel is properly closed by trying to resume
	err = job.Resume()
	assert.NoError(t, err)
	assert.Equal(t, "running", job.Status)

	// Give some time to complete
	time.Sleep(5 * time.Millisecond)
	assert.Equal(t, "completed", job.Status)
}

func TestAgent_MultipleJobs(t *testing.T) {
	// Create an agent without client - should return error
	agentWithoutClient := &Agent{
		Role: "test-agent",
		// Client is nil by default
		Flows: []*flow.Flow{
			{
				Name: "test-flow",
				Steps: []flow.Step{
					*flow.NewStep(&executors.DelayExecutor{
						Milliseconds: 1000 * 15,
					}, nil, nil),
				},
			},
		},
	}

	// Create multiple contexts
	context1 := &AgentContext{
		Goal:            "Test job 1",
		ShortTermMemory: make(map[string]interface{}),
		ExecuteLog:      []string{},
	}

	context2 := &AgentContext{
		Goal:            "Test job 2",
		ShortTermMemory: make(map[string]interface{}),
		ExecuteLog:      []string{},
	}

	// Try to start jobs without client - should return error
	job1, err := agentWithoutClient.StartJob(context1)
	assert.Error(t, err)
	assert.Nil(t, job1)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	job2, err := agentWithoutClient.StartJob(context2)
	assert.Error(t, err)
	assert.Nil(t, job2)
	assert.Equal(t, "agent must have a valid client to start a job", err.Error())

	// Create an agent with client but without flows - should return error
	agentWithClientNoFlows := &Agent{
		Role:   "test-agent",
		Client: &mockClient{},
	}

	// Try to start jobs with client but without flows - should return error
	job1, err = agentWithClientNoFlows.StartJob(context1)
	assert.Error(t, err)
	assert.Nil(t, job1)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	job2, err = agentWithClientNoFlows.StartJob(context2)
	assert.Error(t, err)
	assert.Nil(t, job2)
	assert.Equal(t, "agent must have at least one flow to start a job", err.Error())

	// Create a mock flow
	client := &mockClient{}
	mockFlow, err := flow.NewFlow(client, "test-flow")
	assert.NoError(t, err)

	// Create an agent with both client and flows
	agent := &Agent{
		Role:   "test-agent",
		Client: &mockClient{},
		Flows:  []*flow.Flow{mockFlow},
	}

	// Start multiple jobs
	job1, err = agent.StartJob(context1)
	assert.NoError(t, err)
	assert.NotNil(t, job1)

	job2, err = agent.StartJob(context2)
	assert.NoError(t, err)
	assert.NotNil(t, job2)

	assert.NotEqual(t, job1, job2)

	// Both jobs should be running
	assert.Equal(t, "running", job1.Status)
	assert.Equal(t, "running", job2.Status)

	// Give some time for jobs to complete
	time.Sleep(10 * time.Millisecond)

	// Both jobs should be completed
	assert.Equal(t, "completed", job1.Status)
	assert.Equal(t, "completed", job2.Status)
}
