package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ProtocolVersion = "2025-03-26"
	ServerName      = "novel-mcp-server"
	ServerVersion   = "1.0.0"
)

// SetupRoutes registers MCP Streamable HTTP routes on the gin engine
func SetupRoutes(r *gin.Engine) {
	r.POST("/mcp", handleMCPPost)
	r.GET("/mcp", handleMCPGet)
	r.DELETE("/mcp", handleMCPDelete)
}

// handleMCPPost handles all JSON-RPC requests via POST
func handleMCPPost(c *gin.Context) {
	// Validate Accept header - client may request SSE or JSON
	accept := c.GetHeader("Accept")
	sseRequested := strings.Contains(accept, "text/event-stream")

	// Parse request
	var req JSONRPCRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, JSONRPCResponse{
			JSONRPC: "2.0",
			Error:   &RPCError{Code: -32700, Message: "Parse error"},
		})
		return
	}

	// Process request
	resp := processRequest(req)

	// If client requested SSE stream, respond with SSE
	if sseRequested {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		data, _ := json.Marshal(resp)
		fmt.Fprintf(c.Writer, "event: message\ndata: %s\n\n", data)
		c.Writer.(http.Flusher).Flush()
		return
	}

	// Default: respond with JSON
	c.JSON(http.StatusOK, resp)
}

// handleMCPGet handles SSE stream for server-initiated messages (optional, for notifications)
func handleMCPGet(c *gin.Context) {
	// For Streamable HTTP, GET is used for server-to-client notifications
	// We keep it simple - just return 200 with proper headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Status(http.StatusOK)
}

// handleMCPDelete handles session termination
func handleMCPDelete(c *gin.Context) {
	c.Status(http.StatusOK)
}

// processRequest dispatches JSON-RPC requests to the appropriate handler
func processRequest(req JSONRPCRequest) JSONRPCResponse {
	switch req.Method {
	case "initialize":
		return handleInitialize(req)
	case "notifications/initialized":
		// Client notification, no response needed per spec
		// but we return empty to acknowledge
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID}
	case "tools/list":
		return handleToolsList(req)
	case "tools/call":
		return handleToolsCall(req)
	case "resources/list":
		return handleResourcesList(req)
	case "resources/read":
		return handleResourcesRead(req)
	case "ping":
		return JSONRPCResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]interface{}{}}
	default:
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: -32601, Message: fmt.Sprintf("Method not found: %s", req.Method)},
		}
	}
}

func handleInitialize(req JSONRPCRequest) JSONRPCResponse {
	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": false,
			},
			"resources": map[string]interface{}{
				"subscribe":   false,
				"listChanged": false,
			},
		},
		ServerInfo: &ServerInfo{
			Name:    ServerName,
			Version: ServerVersion,
		},
	}
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func handleToolsList(req JSONRPCRequest) JSONRPCResponse {
	tools := getTools()
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  ListToolsResult{Tools: tools},
	}
}

func handleToolsCall(req JSONRPCRequest) JSONRPCResponse {
	// Parse params
	paramsBytes, err := json.Marshal(req.Params)
	if err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: -32602, Message: "Invalid params"},
		}
	}

	var params CallToolParams
	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: -32602, Message: "Invalid params"},
		}
	}

	// Dispatch to handler
	handler, ok := toolDispatcher[params.Name]
	if !ok {
		return JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &RPCError{Code: -32601, Message: fmt.Sprintf("Unknown tool: %s", params.Name)},
		}
	}

	log.Printf("MCP tool call: %s", params.Name)
	result := handler(params.Arguments)
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
}

func handleResourcesList(req JSONRPCRequest) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  ListResourcesResult{Resources: []Resource{}},
	}
}

func handleResourcesRead(req JSONRPCRequest) JSONRPCResponse {
	return JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Error:   &RPCError{Code: -32602, Message: "No resources available"},
	}
}
