package model

type AgentRequest struct {
	Query string `json:"query" binding:"required"`
}

type AgentResponse struct {
	Answer      string      `json:"answer"`
	Plan        ExecutionPlan `json:"plan"`
	Executions  []ToolExecution `json:"executions"`
	TotalTime   int64       `json:"total_time_ms"`
}

type ExecutionPlan struct {
	Steps []PlanStep `json:"steps"`
}

type PlanStep struct {
	StepID      int                    `json:"step_id"`
	ToolName    string                 `json:"tool_name"`
	Arguments   map[string]interface{} `json:"arguments"`
	Description string                 `json:"description"`
	DependsOn   []int                  `json:"depends_on,omitempty"`
}

type ToolExecution struct {
	StepID    int                    `json:"step_id"`
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
	Result    interface{}            `json:"result"`
	Error     string                 `json:"error,omitempty"`
	Duration  int64                  `json:"duration_ms"`
}

type ToolSearchRequest struct {
	Query string `json:"query" binding:"required"`
	TopK  int    `json:"top_k"`
}

type ToolSearchResult struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Score       float32 `json:"score"`
	Schema      JSONMap `json:"schema"`
}
