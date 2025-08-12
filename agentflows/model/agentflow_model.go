package model

// AgentPlanningData represents the data structure passed to the planning template
type AgentPlanningData struct {
	Role              string     `json:"role"`
	BackStory         string     `json:"backStory"`
	PreferredLanguage string     `json:"preferredLanguage"`
	Goal              string     `json:"goal"`
	AvailableFlows    []FlowInfo `json:"availableFlows"`
}

// FlowInfo represents information about available flows for planning
type FlowInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}
