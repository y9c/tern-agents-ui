package chat

import (
	"context"
	"flag"

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

// Models asks the provider; the running session's server does not expose a
// model list through the uniform Session interface.
func (b opencodeBackend) Models(ctx context.Context, _ agent.Session) []agent.Model {
	models, err := b.provider.ListModels(ctx)
	if err != nil {
		return nil
	}
	return models
}
