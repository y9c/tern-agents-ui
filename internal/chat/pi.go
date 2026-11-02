package chat

import (
	"context"
	"flag"

	agent "github.com/saanuregh/agent-cli-connector"
	"github.com/saanuregh/agent-cli-connector/pi"
)

type piBackend struct{ provider *pi.Provider }

// Pi is the backend for the Pi CLI. It leaves the client name unset: a
// name becomes Pi's originator.
func Pi() backend { return piBackend{provider: pi.New(pi.Options{})} }

func (piBackend) Name() string               { return "Pi" }
func (piBackend) Product() string            { return "Pi" }
func (piBackend) Command() string            { return "tpi" }
func (b piBackend) Provider() agent.Provider { return b.provider }

// DefaultModel is empty, so Pi's config chooses the model.
func (piBackend) DefaultModel() string { return "" }

func (piBackend) Defaults(string) (agent.PermissionMode, agent.Effort) { return "", "" }

func (piBackend) Prepare(cfg *agent.Config) { cfg.ProviderOptions = &pi.SessionOptions{} }

// Flags are Pi's own: the provider, extensions, tools and session
// directory the session starts with. The shared -model flag already
// carries the initial model (as Config.Model), so Pi's model list is
// filled from it here.
func (piBackend) Flags(fs *flag.FlagSet) func(cfg *agent.Config) {
	provider := fs.String("provider", "", "the model provider to use (default: Pi's config)")
	extensions := fs.String("extension", "", "an extension file or directory to load (repeatable)")
	tools := fs.String("tools", "", "an allowed tool to enable (repeatable)")
	sessionDir := fs.String("session-dir", "", "where to store sessions (default: Pi's config)")
	return func(cfg *agent.Config) {
		opts := cfg.ProviderOptions.(*pi.SessionOptions)
		opts.Provider = *provider
		opts.Extensions = splitList(*extensions)
		opts.Tools = splitList(*tools)
		opts.SessionDir = *sessionDir
		if cfg.Model != "" {
			opts.Models = splitList(cfg.Model)
		}
	}
}

func (b piBackend) Start(ctx context.Context, cfg agent.Config) (agent.Session, error) {
	return b.provider.Start(ctx, cfg)
}

func (piBackend) Warm(context.Context, agent.Config) {}
func (piBackend) Close()                             {}

// Efforts include the common reasoning levels; Pi accepts any harness
// native string.
func (piBackend) Efforts() []agent.Effort {
	return []agent.Effort{agent.EffortNone, agent.EffortMinimal, agent.EffortLow, agent.EffortMedium,
		agent.EffortHigh, agent.EffortXHigh, agent.EffortMax}
}

// Has reports the features wired for Pi. Shell commands dispatch on the
// *pi.Session type (pi.Bash, see shellLocked); the rest stay against
// Claude's and Codex's session types and are not offered.
func (piBackend) Has(f feature) bool {
	switch f {
	case featShell:
		return true
	}
	return false
}

// SessionExists reads one turn of the stored session.
func (b piBackend) SessionExists(ctx context.Context, id, cwd string) error {
	_, err := b.provider.ReadHistory(ctx, id, agent.ReadHistoryOptions{Cwd: cwd, Limit: 1})
	return err
}

// FileSuggestions has no Pi equivalent in the connector; the @-mention
// file picker is not offered.
func (piBackend) FileSuggestions(context.Context, agent.Session, string, string) ([]string, error) {
	return nil, agent.ErrUnsupported
}

// Models asks the provider; the running session's server does not expose a
// model list through the uniform Session interface.
func (b piBackend) Models(ctx context.Context, _ agent.Session) []agent.Model {
	models, err := b.provider.ListModels(ctx)
	if err != nil {
		return nil
	}
	return models
}
