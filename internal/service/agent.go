package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"mcp-agent/internal/config"
	"mcp-agent/internal/model"
	"mcp-agent/internal/repository"
	"mcp-agent/pkg/embedding"
	"mcp-agent/pkg/llm"
	"mcp-agent/pkg/logger"
	"mcp-agent/pkg/vectordb"

	"go.uber.org/zap"
)

type AgentService struct {
	toolRepo      *repository.ToolRepository
	toolSvc       *ToolService
	llmClient     *llm.Client
	embClient     *embedding.Client
	vectorDB      vectordb.VectorDB
	cfg           config.Config
}

func NewAgentService(
	toolRepo *repository.ToolRepository,
	toolSvc *ToolService,
	llmClient *llm.Client,
	embClient *embedding.Client,
	vectorDB vectordb.VectorDB,
	cfg config.Config,
) *AgentService {
	return &AgentService{
		toolRepo:  toolRepo,
		toolSvc:   toolSvc,
		llmClient: llmClient,
		embClient: embClient,
		vectorDB:  vectorDB,
		cfg:       cfg,
	}
}

func (s *AgentService) SearchTools(query string, topK int) ([]model.ToolSearchResult, error) {
	if topK <= 0 {
		topK = 5
	}

	queryVec, err := s.embClient.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	results, err := s.vectorDB.Search(queryVec, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to search vector db: %w", err)
	}

	searchResults := make([]model.ToolSearchResult, 0, len(results))
	for _, r := range results {
		name, _ := r.Metadata["name"].(string)
		desc, _ := r.Metadata["description"].(string)
		schemaData, _ := r.Metadata["schema"].(map[string]interface{})

		searchResults = append(searchResults, model.ToolSearchResult{
			Name:        name,
			Description: desc,
			Score:       r.Score,
			Schema:      schemaData,
		})
	}

	return searchResults, nil
}

func (s *AgentService) Execute(query string, callerID int64, callerName string) (*model.AgentResponse, error) {
	start := time.Now()

	// Step 1: 工具检索
	tools, err := s.SearchTools(query, 5)
	if err != nil {
		return nil, fmt.Errorf("tool search failed: %w", err)
	}

	if len(tools) == 0 {
		return nil, fmt.Errorf("no relevant tools found")
	}

	// Step 2: 生成执行计划
	plan, err := s.generatePlan(query, tools)
	if err != nil {
		return nil, fmt.Errorf("plan generation failed: %w", err)
	}

	// Step 3: 执行计划
	executions, err := s.executePlan(plan, callerID, callerName)
	if err != nil {
		logger.Error("plan execution failed", zap.Error(err))
	}

	// Step 4: 生成最终回答
	answer, err := s.generateAnswer(query, executions)
	if err != nil {
		return nil, fmt.Errorf("answer generation failed: %w", err)
	}

	totalTime := time.Since(start).Milliseconds()

	return &model.AgentResponse{
		Answer:     answer,
		Plan:       *plan,
		Executions: executions,
		TotalTime:  totalTime,
	}, nil
}

func (s *AgentService) generatePlan(query string, tools []model.ToolSearchResult) (*model.ExecutionPlan, error) {
	toolsDesc := s.formatToolsForPrompt(tools)

	prompt := fmt.Sprintf(`You are a task planner. Given a user query and available tools, generate a step-by-step execution plan.

User Query: %s

Available Tools:
%s

Generate a JSON execution plan with the following structure:
{
  "steps": [
    {
      "step_id": 1,
      "tool_name": "tool_name",
      "arguments": {"arg1": "value1"},
      "description": "what this step does",
      "depends_on": []
    }
  ]
}

Rules:
1. Only use tools from the available tools list
2. Each step must have a unique step_id starting from 1
3. If a step depends on previous steps, list their step_ids in depends_on
4. Keep the plan minimal and efficient
5. Return ONLY the JSON, no explanation

Plan:`, query, toolsDesc)

	messages := []llm.Message{
		{Role: "user", Content: prompt},
	}

	response, err := s.llmClient.Chat(messages, s.cfg.LLM.Model.Planner)
	if err != nil {
		return nil, err
	}

	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		response = strings.TrimSuffix(response, "```")
		response = strings.TrimSpace(response)
	}

	var plan model.ExecutionPlan
	if err := json.Unmarshal([]byte(response), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan: %w", err)
	}

	return &plan, nil
}

func (s *AgentService) executePlan(plan *model.ExecutionPlan, callerID int64, callerName string) ([]model.ToolExecution, error) {
	executions := make([]model.ToolExecution, 0, len(plan.Steps))
	results := make(map[int]interface{})

	for _, step := range plan.Steps {
		// 检查依赖
		for _, depID := range step.DependsOn {
			if _, exists := results[depID]; !exists {
				return executions, fmt.Errorf("step %d depends on step %d which failed", step.StepID, depID)
			}
		}

		start := time.Now()
		result, err := s.toolSvc.CallTool(step.ToolName, step.Arguments, callerID, callerName)
		duration := time.Since(start).Milliseconds()

		execution := model.ToolExecution{
			StepID:    step.StepID,
			ToolName:  step.ToolName,
			Arguments: step.Arguments,
			Duration:  duration,
		}

		if err != nil {
			execution.Error = err.Error()
			logger.Error("step execution failed",
				zap.Int("step_id", step.StepID),
				zap.String("tool", step.ToolName),
				zap.Error(err))
		} else {
			execution.Result = result
			results[step.StepID] = result
		}

		executions = append(executions, execution)
	}

	return executions, nil
}

func (s *AgentService) generateAnswer(query string, executions []model.ToolExecution) (string, error) {
	executionSummary := s.formatExecutionsForPrompt(executions)

	prompt := fmt.Sprintf(`You are a helpful assistant. Based on the user's query and the tool execution results, provide a clear and concise answer.

User Query: %s

Tool Execution Results:
%s

Provide a natural language answer that directly addresses the user's query. Be concise and helpful.

Answer:`, query, executionSummary)

	messages := []llm.Message{
		{Role: "user", Content: prompt},
	}

	return s.llmClient.Chat(messages, s.cfg.LLM.Model.Final)
}

func (s *AgentService) formatToolsForPrompt(tools []model.ToolSearchResult) string {
	var sb strings.Builder
	for i, tool := range tools {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, tool.Name))
		sb.WriteString(fmt.Sprintf("   Description: %s\n", tool.Description))
		if tool.Schema != nil {
			schemaJSON, _ := json.MarshalIndent(tool.Schema, "   ", "  ")
			sb.WriteString(fmt.Sprintf("   Schema: %s\n", string(schemaJSON)))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (s *AgentService) formatExecutionsForPrompt(executions []model.ToolExecution) string {
	var sb strings.Builder
	for _, exec := range executions {
		sb.WriteString(fmt.Sprintf("Step %d - %s:\n", exec.StepID, exec.ToolName))
		if exec.Error != "" {
			sb.WriteString(fmt.Sprintf("  Error: %s\n", exec.Error))
		} else {
			resultJSON, _ := json.MarshalIndent(exec.Result, "  ", "  ")
			sb.WriteString(fmt.Sprintf("  Result: %s\n", string(resultJSON)))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func (s *AgentService) IndexTools() error {
	tools, err := s.toolRepo.ListEnabled()
	if err != nil {
		return fmt.Errorf("failed to list tools: %w", err)
	}

	for _, tool := range tools {
		text := fmt.Sprintf("%s: %s", tool.Name, tool.Description)
		vector, err := s.embClient.Embed(text)
		if err != nil {
			logger.Error("failed to embed tool", zap.String("tool", tool.Name), zap.Error(err))
			continue
		}

		metadata := map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"schema":      tool.Schema,
		}

		if err := s.vectorDB.Insert(tool.Name, vector, metadata); err != nil {
			logger.Error("failed to insert tool vector", zap.String("tool", tool.Name), zap.Error(err))
			continue
		}

		logger.Info("indexed tool", zap.String("tool", tool.Name))
	}

	return nil
}
