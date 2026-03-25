package native

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"mcp-agent/internal/model"
)

// Executor 执行内置原生工具
type Executor struct {
	httpClient *http.Client
}

func NewExecutor() *Executor {
	return &Executor{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Execute 根据 NativeTemplate 分发执行
func (e *Executor) Execute(tool *model.Tool, args map[string]interface{}) (interface{}, error) {
	tpl, _ := tool.NativeConfig["template"].(string)
	switch model.NativeTemplate(tpl) {
	case model.NativeTplHTTPGet:
		return e.httpGet(tool, args)
	case model.NativeTplHTTPPost:
		return e.httpPost(tool, args)
	case model.NativeTplReadDoc:
		return e.readDoc(tool, args)
	default:
		return nil, fmt.Errorf("unknown native template: %q", tpl)
	}
}

// ---------------------------------------------------------------------------
// http_get
// native_config: { "template": "http_get", "url": "https://...", "headers": {"X-Key": "val"} }
// args 中的 key 会替换 url 里的 {key} 占位符
// ---------------------------------------------------------------------------
func (e *Executor) httpGet(tool *model.Tool, args map[string]interface{}) (interface{}, error) {
	rawURL, _ := tool.NativeConfig["url"].(string)
	if rawURL == "" {
		return nil, fmt.Errorf("native_config.url is required for http_get")
	}
	resolvedURL := resolvePlaceholders(rawURL, args)

	req, err := http.NewRequest(http.MethodGet, resolvedURL, nil)
	if err != nil {
		return nil, err
	}
	addHeaders(req, tool.NativeConfig)

	return e.doRequest(req)
}

// ---------------------------------------------------------------------------
// http_post
// native_config: { "template": "http_post", "url": "https://...", "headers": {}, "body_template": "{\"q\": \"{query}\"}" }
// ---------------------------------------------------------------------------
func (e *Executor) httpPost(tool *model.Tool, args map[string]interface{}) (interface{}, error) {
	rawURL, _ := tool.NativeConfig["url"].(string)
	if rawURL == "" {
		return nil, fmt.Errorf("native_config.url is required for http_post")
	}
	resolvedURL := resolvePlaceholders(rawURL, args)

	var bodyReader io.Reader
	if bodyTpl, ok := tool.NativeConfig["body_template"].(string); ok && bodyTpl != "" {
		body := resolvePlaceholders(bodyTpl, args)
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(http.MethodPost, resolvedURL, bodyReader)
	if err != nil {
		return nil, err
	}
	if bodyReader != nil {
		contentType := "application/json"
		if ct, ok := tool.NativeConfig["content_type"].(string); ok && ct != "" {
			contentType = ct
		}
		req.Header.Set("Content-Type", contentType)
	}
	addHeaders(req, tool.NativeConfig)

	return e.doRequest(req)
}

// ---------------------------------------------------------------------------
// read_doc
// native_config: { "template": "read_doc", "source": "url|text", "url": "https://...", "content": "静态文本" }
// ---------------------------------------------------------------------------
func (e *Executor) readDoc(tool *model.Tool, args map[string]interface{}) (interface{}, error) {
	source, _ := tool.NativeConfig["source"].(string)
	switch source {
	case "url":
		rawURL, _ := tool.NativeConfig["url"].(string)
		if rawURL == "" {
			return nil, fmt.Errorf("native_config.url is required when source=url")
		}
		resolvedURL := resolvePlaceholders(rawURL, args)
		req, err := http.NewRequest(http.MethodGet, resolvedURL, nil)
		if err != nil {
			return nil, err
		}
		return e.doRequest(req)
	case "text":
		content, _ := tool.NativeConfig["content"].(string)
		if content == "" {
			return nil, fmt.Errorf("native_config.content is required when source=text")
		}
		return map[string]interface{}{"content": content}, nil
	default:
		return nil, fmt.Errorf("native_config.source must be 'url' or 'text'")
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func (e *Executor) doRequest(req *http.Request) (interface{}, error) {
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	return map[string]interface{}{
		"status": resp.StatusCode,
		"body":   string(body),
	}, nil
}

// resolvePlaceholders 将字符串中的 {key} 替换为 args[key]
func resolvePlaceholders(s string, args map[string]interface{}) string {
	for k, v := range args {
		s = strings.ReplaceAll(s, "{"+k+"}", fmt.Sprintf("%v", v))
	}
	return s
}

func addHeaders(req *http.Request, cfg model.JSONMap) {
	headers, ok := cfg["headers"].(map[string]interface{})
	if !ok {
		return
	}
	for k, v := range headers {
		req.Header.Set(k, fmt.Sprintf("%v", v))
	}
}
