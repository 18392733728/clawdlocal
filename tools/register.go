package tools

import (
	"context"
	"clawdlocal/core"
)

// RegisterAllTools registers all built-in tools with the agent
func RegisterAllTools(agent *core.Agent) {
	// File operations
	agent.ToolManager.RegisterTool(&core.Tool{
		Name:        "file_read",
		Description: "Read the contents of a file from the filesystem",
		Parameters:  (&FileReadTool{}).Parameters(),
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return (&FileReadTool{}).Execute(ctx, args)
		},
	})
	
	agent.ToolManager.RegisterTool(&core.Tool{
		Name:        "file_write",
		Description: "Write content to a file in the filesystem",
		Parameters:  (&FileWriteTool{}).Parameters(),
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return (&FileWriteTool{}).Execute(ctx, args)
		},
	})
	
	agent.ToolManager.RegisterTool(&core.Tool{
		Name:        "file_list",
		Description: "List contents of a directory",
		Parameters:  (&FileListTool{}).Parameters(),
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return (&FileListTool{}).Execute(ctx, args)
		},
	})
	
	// Network operations
	agent.ToolManager.RegisterTool(&core.Tool{
		Name:        "network_request",
		Description: "Make HTTP requests to external services",
		Parameters:  (&NetworkRequestTool{}).Parameters(),
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return (&NetworkRequestTool{}).Execute(ctx, args)
		},
	})
	
	// Database operations
	agent.ToolManager.RegisterTool(&core.Tool{
		Name:        "database_query",
		Description: "Execute SQL queries against a database",
		Parameters:  (&DatabaseQueryTool{}).Parameters(),
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return (&DatabaseQueryTool{}).Execute(ctx, args)
		},
	})
}