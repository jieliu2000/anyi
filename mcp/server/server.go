package server

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jieliu2000/anyi/flow"
)

type MCPServer struct {
	ginEngine *gin.Engine
	flows     map[string]flow.Flow
}

func New() *MCPServer {
	r := gin.Default()
	return &MCPServer{
		ginEngine: r,
		flows:     make(map[string]flow.Flow),
	}
}

func (s *MCPServer) Start(addr string) error {
	// Setup basic API routes
	s.ginEngine.POST("/api/flow/:name/execute", s.handleExecuteFlow)
	s.ginEngine.GET("/api/flow/list", s.handleListFlows)

	// Start HTTP server
	return s.ginEngine.Run(addr)
}

func (s *MCPServer) RegisterFlow(name string, f flow.Flow) {
	s.flows[name] = f
}

func (s *MCPServer) handleExecuteFlow(c *gin.Context) {
	name := c.Param("name")
	f, exists := s.flows[name]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "flow not found"})
		return
	}

	var input map[string]interface{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert input to JSON string
	inputJSON, err := json.Marshal(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Run flow with input
	result, err := f.RunWithInput(string(inputJSON))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (s *MCPServer) handleListFlows(c *gin.Context) {
	flowNames := make([]string, 0, len(s.flows))
	for name := range s.flows {
		flowNames = append(flowNames, name)
	}
	c.JSON(http.StatusOK, gin.H{"flows": flowNames})
}
