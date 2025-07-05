package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ibra86/k8s-controller-patterns/pkg/api"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func listFrontendPagesHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if api.FrontendAPI == nil {
		return mcp.NewToolResultText("FrontendAPI is not initialized"), nil
	}
	docs, err := api.FrontendAPI.ListFrontendPagesRaw(ctx)
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error listing FrontendPages: %v", err)), nil
	}
	jsonBytes, err := json.MarshalIndent(docs, "", " ")
	if err != nil {
		return mcp.NewToolResultText(fmt.Sprintf("Error marshaling result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func createFrontendPageHandler(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// TODO
	return mcp.NewToolResultText("Created FrontendPage (stub)"), nil
}

func NewMCPServer(serverName, version string) *server.MCPServer {
	s := server.NewMCPServer(
		serverName,
		version,
		server.WithToolCapabilities(true),
		server.WithLogging(),
		server.WithRecovery(),
	)

	listTool := mcp.NewTool(
		"list_frontendpages",
		mcp.WithDescription("List all FrontendPage resources"),
	)
	createTool := mcp.NewTool(
		"create_frontendpages",
		mcp.WithDescription("Create a new FrontendPage resources"),
		mcp.WithString("name", mcp.Description("Name of the FrontendPage")),
		mcp.WithString("contents", mcp.Description("HTML contents")),
		mcp.WithString("image", mcp.Description("Container image")),
		mcp.WithString("replicas", mcp.Description("Number of replicas")),
	)
	// TODO: update, delete

	s.AddTool(listTool, listFrontendPagesHandler)
	s.AddTool(createTool, createFrontendPageHandler)
	return s
}
