package chat

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"sort"

	agent "github.com/saanuregh/agent-cli-connector"
	"github.com/saanuregh/agent-cli-connector/opencode"
)

type opencodeBackend struct{ provider *opencode.Provider }

// OpenCode is the backend for the OpenCode CLI. It leaves the client name
// unset: a name becomes OpenCode's originator.
func OpenCode() backend { return opencodeBackend{provider: opencode.New(opencode.Options{})} }

func (opencodeBackend) Name() string               { return "OpenCode" }
func (opencodeBackend) Product() string            { return "OpenCode" }
func (opencodeBackend) Command() string            { return "topencode" }
func (b opencodeBackend) Provider() agent.Provider { return b.provider }

// DefaultModel is empty, so the session's server config chooses the model.
func (opencodeBackend) DefaultModel() string { return "" }

func (opencodeBackend) Defaults(string) (agent.PermissionMode, agent.Effort) { return "", "" }

func (opencodeBackend) Prepare(cfg *agent.Config) { cfg.ProviderOptions = &opencode.SessionOptions{} }

// Flags are OpenCode's own: the agent to run turns with.
func (opencodeBackend) Flags(fs *flag.FlagSet) func(cfg *agent.Config) {
	agentName := fs.String("agent", "", "the agent to run turns with, such as build or plan (default: OpenCode settings)")
	return func(cfg *agent.Config) {
		opts := cfg.ProviderOptions.(*opencode.SessionOptions)
		opts.Agent = *agentName
	}
}

func (b opencodeBackend) Start(ctx context.Context, cfg agent.Config) (agent.Session, error) {
	return b.provider.Start(ctx, cfg)
}

func (opencodeBackend) Warm(context.Context, agent.Config) {}
func (opencodeBackend) Close()                             {}

// Efforts include the common reasoning levels; OpenCode accepts any
// harness-native string.
func (opencodeBackend) Efforts() []agent.Effort {
	return []agent.Effort{agent.EffortNone, agent.EffortMinimal, agent.EffortLow, agent.EffortMedium,
		agent.EffortHigh, agent.EffortXHigh, agent.EffortMax}
}

// Has reports the features wired for OpenCode. The commands behind them
// (see opencode_extras.go) dispatch on the *opencode.Session type; the
// rest stay written against Claude's and Codex's *Session types and are
// not offered.
func (opencodeBackend) Has(f feature) bool {
	switch f {
	case featAgents, featPermissionRules, featSideQuestion, featTurnDiff, featExport, featShell:
		return true
	}
	return false
}

// SessionExists reads one turn of the stored session.
func (b opencodeBackend) SessionExists(ctx context.Context, id, cwd string) error {
	_, err := b.provider.ReadHistory(ctx, id, agent.ReadHistoryOptions{Cwd: cwd, Limit: 1})
	return err
}

// FileSuggestions has no OpenCode equivalent in the connector; the @-mention
// file picker is not offered.
func (opencodeBackend) FileSuggestions(context.Context, agent.Session, string, string) ([]string, error) {
	return nil, agent.ErrUnsupported
}

// Models lists the models to offer. The provider's ListModels reflects the
// running OpenCode server, which does not enumerate a provider's models until
// they are exercised in a session (a cold /api/model returns none), so the
// configured set from opencode.json is the source of truth and is merged in.
func (b opencodeBackend) Models(ctx context.Context, _ agent.Session) []agent.Model {
	listed, err := b.provider.ListModels(ctx)
	if err != nil {
		listed = nil
	}
	return mergeModels(listed, opencodeConfigModels())
}

// opencodeConfigModels reads ~/.config/opencode/opencode.json and returns one
// agent.Model per configured provider model, as "provider/model".
func opencodeConfigModels() []agent.Model {
	raw, err := os.ReadFile(opencodeConfigPath())
	if err != nil {
		return nil
	}
	var cfg struct {
		Model    string `json:"model"`
		Provider map[string]struct {
			Models map[string]struct {
				Name    string `json:"name"`
				Limit   struct {
					Context int64 `json:"context"`
				} `json:"limit"`
				Modalities *struct {
					Input []string `json:"input"`
				} `json:"modalities"`
			} `json:"models"`
		} `json:"provider"`
	}
	if json.Unmarshal(raw, &cfg) != nil {
		return nil
	}
	var out []agent.Model
	for pid, prov := range cfg.Provider {
		for mid, md := range prov.Models {
			name := md.Name
			if name == "" {
				name = mid
			}
			m := agent.Model{ID: pid + "/" + mid, DisplayName: name, Provider: pid, ContextWindow: md.Limit.Context}
			if cfg.Model == m.ID {
				m.IsDefault = true
			}
			if mdl := md.Modalities; mdl != nil {
				m.InputModalities = mdl.Input
			}
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Provider != out[j].Provider {
			return out[i].Provider < out[j].Provider
		}
		return out[i].ID < out[j].ID
	})
	return out
}

// opencodeConfigPath is the user config OpenCode loads its providers from.
func opencodeConfigPath() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "opencode", "opencode.json")
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "opencode", "opencode.json")
}

// mergeModels returns a and b deduped by ID. A missing window or modality on
// one entry is filled from the other; the config set fills the provider's
// cold enumeration.
func mergeModels(a, b []agent.Model) []agent.Model {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	var out []agent.Model
	for _, m := range append(append([]agent.Model{}, a...), b...) {
		i := indexModel(out, m.ID)
		if i < 0 {
			out = append(out, m)
			continue
		}
		if m.ContextWindow != 0 && out[i].ContextWindow == 0 {
			out[i].ContextWindow = m.ContextWindow
		}
		if len(m.InputModalities) > 0 && len(out[i].InputModalities) == 0 {
			out[i].InputModalities = m.InputModalities
		}
		if m.IsDefault {
			out[i].IsDefault = true
		}
	}
	return out
}

func indexModel(ms []agent.Model, id string) int {
	for i, m := range ms {
		if m.ID == id {
			return i
		}
	}
	return -1
}
