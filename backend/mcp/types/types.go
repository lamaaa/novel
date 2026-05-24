package types

// CallToolResult is the result of a tool call
type CallToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ContentItem represents a piece of content in a tool result
type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
