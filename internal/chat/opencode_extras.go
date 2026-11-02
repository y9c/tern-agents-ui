package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/saanuregh/agent-cli-connector/opencode"
)

// OpenCode's own controls that the shared API lacks: subagents, permission
// rules, a turn diff, a side question and shell commands. Each helper runs
// off the lock and shows its card or toast with mu held, as the Claude and
// Codex equivalents do.

// opencodeAgents lists the session's OpenCode agents as a card.
func (a *app) opencodeAgents(os *opencode.Session) {
	gen := a.gen
	a.async("Subagents", func(ctx context.Context) error {
		agents, err := os.Agents(ctx)
		if err != nil {
			return err
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		if gen != a.gen {
			return nil
		}
		if len(agents) == 0 {
			a.toastLocked("info", "No subagents are defined")
			return nil
		}
		var b strings.Builder
		b.WriteString("| Agent | Mode | What it does |\n| --- | --- | --- |\n")
		for _, ag := range agents {
			fmt.Fprintf(&b, "| `%s` | %s | %s |\n", ag.Name, firstNonEmpty(ag.Mode, "all"), strings.ReplaceAll(firstLine(ag.Description), "|", "\\|"))
		}
		a.addLocal(localCard{Title: "Subagents", Text: b.String()})
		return nil
	})
}

// opencodePermissions shows the session's permission rules as a card.
func (a *app) opencodePermissions(os *opencode.Session) {
	gen := a.gen
	a.async("Permission rules", func(ctx context.Context) error {
		rules, err := os.PermissionRules(ctx)
		if err != nil {
			return err
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		if gen != a.gen {
			return nil
		}
		var b strings.Builder
		if len(rules) == 0 {
			b.WriteString("No permission rules.")
		} else {
			b.WriteString("| Effect | Action | Resource |\n| --- | --- | --- |\n")
		}
		for _, r := range rules {
			fmt.Fprintf(&b, "| %s | `%s` | `%s` |\n", r.Effect, r.Action, r.Resource)
		}
		a.addLocal(localCard{Title: "Permission rules", Text: b.String()})
		return nil
	})
}

// opencodeDiff shows what the session's current or last turn changed.
func (a *app) opencodeDiff(os *opencode.Session) {
	gen := a.gen
	a.async("Diff", func(ctx context.Context) error {
		raw, err := os.Diff(ctx)
		if err != nil {
			return err
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		if gen != a.gen {
			return nil
		}
		if len(strings.TrimSpace(string(raw))) == 0 {
			a.toastLocked("info", "No changes in this turn yet")
			return nil
		}
		a.addLocal(localCard{Title: "Changes this turn", Text: prettyJSON(raw)})
		return nil
	})
}

// opencodeBtw asks a side question that stays out of the conversation.
func (a *app) opencodeBtw(os *opencode.Session, arg string) {
	gen := a.gen
	a.async("Side question", func(ctx context.Context) error {
		answer, err := os.SideQuestion(ctx, arg)
		if err != nil {
			return err
		}
		a.mu.Lock()
		defer a.mu.Unlock()
		if gen != a.gen {
			return nil
		}
		a.addLocal(localCard{Title: "Side question: " + firstLine(arg), Text: firstNonEmpty(answer, "_No answer._"), Tone: "info"})
		return nil
	})
}

// opencodeShell runs a shell command in the conversation, as OpenCode does
// for its own input; output arrives as items.
func (a *app) opencodeShell(os *opencode.Session, cmd string, restore func()) {
	gen := a.gen
	go func() {
		ctx, cancel := context.WithTimeout(a.ctx, 2*time.Minute)
		defer cancel()
		if err := os.Shell(ctx, cmd); err != nil && !errors.Is(err, context.Canceled) {
			a.mu.Lock()
			defer a.mu.Unlock()
			a.toastLocked("error", "Shell: "+err.Error())
			if gen == a.gen {
				restore()
			}
		}
	}()
}
