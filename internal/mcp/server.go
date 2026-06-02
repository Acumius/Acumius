package mcp

import (
	"context"

	"github.com/google/uuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/Acumius/Acumius/internal/memory"
	"github.com/Acumius/Acumius/internal/policy"
)

// Server wraps the MCP server.
type Server struct {
	mcpServer *server.MCPServer
	memStore  memory.Store
	evaluator *policy.Evaluator
}

// NewServer creates a new Acumius MCP Server.
func NewServer(memStore memory.Store, evaluator *policy.Evaluator) *Server {
	s := server.NewMCPServer("acumius-mcp", "1.0.0", server.WithResourceCapabilities(true, true), server.WithPromptCapabilities(true))

	srv := &Server{
		mcpServer: s,
		memStore:  memStore,
		evaluator: evaluator,
	}

	srv.registerTools()
	return srv
}

func (s *Server) registerTools() {
	// memory_retrieve Tool
	memoryRetrieveTool := mcp.NewTool("memory_retrieve",
		mcp.WithDescription("Retrieve a memory by ID"),
		mcp.WithString("id", mcp.Required(), mcp.Description("The UUID of the memory")),
	)
	s.mcpServer.AddTool(memoryRetrieveTool, s.handleMemoryRetrieve)
}

func (s *Server) handleMemoryRetrieve(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	idStr, err := request.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError("missing or invalid id argument"), nil
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		return mcp.NewToolResultError("invalid uuid format"), nil
	}

	mem, err := s.memStore.Retrieve(ctx, id, memory.RetrieveOpts{})
	if err != nil {
		return mcp.NewToolResultError("memory not found: " + err.Error()), nil
	}

	// Just return a summary for now
	return mcp.NewToolResultText("Retrieved memory successfully (namespace: " + mem.Namespace + ")"), nil
}

// ServeStdio starts the MCP server over standard I/O
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}

// SSEServer returns an SSEServer instance
func (s *Server) SSEServer() *server.SSEServer {
	return server.NewSSEServer(s.mcpServer)
}
