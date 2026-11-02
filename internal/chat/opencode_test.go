package chat

import (
	"os"
	"path/filepath"
	"testing"

	agent "github.com/saanuregh/agent-cli-connector"
)

// TestOpenCodeConfigModels parses a provider/model config from opencode.json.
// OpenCode's cold /api/model returns no models for a provider, so the app
// lists what the config declares; this pins that behavior.
func TestOpenCodeConfigModels(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "opencode"), 0o700); err != nil {
		t.Fatal(err)
	}
	cfg := `{
		"model": "hpc-gateway/deepseek-v4-flash",
		"provider": {
			"hpc-gateway": {
				"models": {
					"deepseek-v4-flash": {"name": "DeepSeek", "limit": {"context": 1048576}},
					"qwen3.8-27B": {"modalities": {"input": ["text", "image"]}}
				}
			},
			"other": {"models": {"mini": {"name": "Mini"}}}
		}
	}`
	if err := os.WriteFile(filepath.Join(dir, "opencode", "opencode.json"), []byte(cfg), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", dir)

	models := opencodeConfigModels()
	if len(models) != 3 {
		t.Fatalf("got %d models, want 3: %+v", len(models), models)
	}
	byID := map[string]agent.Model{}
	for _, m := range models {
		byID[m.ID] = m
	}
	if got := byID["hpc-gateway/deepseek-v4-flash"]; !got.IsDefault || got.ContextWindow != 1048576 || got.DisplayName != "DeepSeek" {
		t.Fatalf("deepseek = %+v, want default with 1M window", got)
	}
	if got := byID["hpc-gateway/qwen3.8-27B"]; len(got.InputModalities) != 2 {
		t.Fatalf("qwen modalities = %+v, want text+image", got.InputModalities)
	}
	if got := byID["other/mini"]; got.DisplayName != "Mini" {
		t.Fatalf("mini = %+v, want name populated", got)
	}
}

// TestMergeModels fills a cold provider list from the config set.
func TestMergeModels(t *testing.T) {
	a := []agent.Model{ // provider's cold enumeration: default only
		{ID: "hpc-gateway/deepseek-v4-flash", Provider: "hpc-gateway", IsDefault: true, ContextWindow: 1048576},
	}
	b := []agent.Model{ // config: two models, one missing a window
		{ID: "hpc-gateway/deepseek-v4-flash", Provider: "hpc-gateway", IsDefault: true},
		{ID: "hpc-gateway/qwen3.8-27B", Provider: "hpc-gateway"},
	}
	got := mergeModels(a, b)
	if len(got) != 2 {
		t.Fatalf("got %d models, want 2: %+v", len(got), got)
	}
	for _, m := range got {
		if m.ID == "hpc-gateway/deepseek-v4-flash" && (!m.IsDefault || m.ContextWindow != 1048576) {
			t.Fatalf("deepseek merged = %+v, want default+window from a", m)
		}
	}
}
