# Claude Code, Codex, OpenCode and Pi

`tclaude` drives the `claude` CLI, `tcodex` drives the `codex` CLI,
`topencode` drives the `opencode` CLI and `tpi` drives the `pi` CLI. All
four share the same chat; this page covers what differs. Hermes is also a
supported native Tern agent, served by the separate hermes-for-tern
frontend rather than a Go backend built here (see below).

## Command-line flags

```text
tclaude [-model ID] [-mode MODE] [-effort LEVEL] [-resume ID | -continue] [-cwd DIR]
        [-add-dir DIR]... [-system-prompt TEXT] [-append-system-prompt TEXT]
        [-mcp-config FILE] [-no-save] [-title TITLE] [-log FILE [-debug]]
        [-agent NAME] [-fallback-model ID] [-max-turns N] [-max-budget USD]
        [-allowed-tools LIST] [-disallowed-tools LIST] [-allow-bypass]

tcodex  [the shared flags above, up to -debug]
        [-service-tier TIER] [-personality NAME] [-reasoning-summary LEVEL]

topencode [the shared flags above, up to -debug]
          [-agent NAME]   # the agent to run turns with, such as build or plan

tpi [the shared flags above, up to -debug]
    [-provider NAME]    # the model provider to use (default: Pi's config)
    [-extension FILE]   # an extension file or directory to load (repeatable)
    [-tools TOOL]       # an allowed tool to enable (repeatable)
    [-session-dir DIR]  # where to store sessions (default: Pi's config)
```

- `-help` describes each flag.
- `-mcp-config` takes Claude Code's `.mcp.json` format.
- With `-log FILE -debug`, tclaude also writes Claude Code's own debug log
  to `FILE.claude`. Debug logs can contain conversation text.

### Default model, effort and mode

- `tclaude` starts on `claude-haiku-5-5` to keep a new chat cheap
  (`DefaultModel` in `internal/chat/claude.go`). Without `-mode` or
  `-effort`, your Claude Code settings decide.
- `tcodex` passes no model or effort, so `~/.codex/config.toml` decides.
- `topencode` and `tpi` pass no model or effort, so their configs decide.

## Pi

### Sharing conversations with `pi`

Conversations are Pi sessions, run through Pi's RPC mode with your Pi
config, `AGENTS.md`, MCP servers and skills. `tpi` passes no model by
default; `-model` (and `-provider` for the provider) selects one.

- `tpi -provider NAME` chooses the model provider.
- `-extension FILE` loads an extension; `-tools TOOL` limits the enabled
  tools; `-session-dir DIR` stores sessions elsewhere.

Pi sends background-task events like the other agents, so `/tasks` lists
and stops them. `!command` runs a shell command in the conversation
through Pi's Bash.

### What tpi leaves out

/agents, /permissions, /btw, /diff, /export, /review, /goal, /plan,
/usage, /reload, /cd and Ctrl+B have no Pi equivalent here.

## Claude Code

### Sharing conversations with `claude`

Conversations are ordinary Claude Code sessions in `~/.claude/projects/`,
with your settings, `CLAUDE.md`, hooks, MCP servers and skills.

- A `tclaude` chat shows in `claude --resume` and `claude --continue`, and
  `claude --resume <id>` opens it. Sessions are recorded with the
  entrypoint `tclaude`, because Claude Code hides SDK-tagged sessions from
  its list.
- A `claude` chat shows in `/resume` here, and `tclaude -resume <id>` or
  `-continue` opens it.
- Don't keep one session open in both at once. Each appends to the same
  file without seeing the other's messages.

### Claude-only commands

| Command | Does |
| --- | --- |
| `/agents` | List the subagents Claude can use |
| `/permissions` | Show the permission rules |
| `/usage` | Plan usage |
| `/reload` | Reload skills, plugins and output styles |
| `/cd <folder>` | Move the conversation to another folder |
| `/mcp login\|logout <server>` | Sign in to or out of an MCP server |
| `/btw <question>` | A side question that stays out of the conversation |
| `/plan` | Show the current plan file |
| Ctrl+B | Move a running command to the background |

## Codex

### Sharing conversations with `codex`

Conversations are Codex threads, run through `codex app-server` with your
`~/.codex` config, `AGENTS.md` and MCP servers. Whether `codex resume` lists
threads started from `tcodex` is untested.

### Codex-only commands

| Command | Does |
| --- | --- |
| `/review [uncommitted\|branch <base>\|commit <sha> [title]\|instructions]` | A code review turn |
| `/goal [objective\|pause\|resume\|clear]` | The goal Codex works toward |
| `!command` | Run a shell command in the conversation |
| `/diff` | What the current or last turn changed |
| `/tasks clean` | Stop every background terminal (`/tasks` lists them; Enter stops one) |
| `/archive`, `/unarchive` | Archive or restore the thread |
| `/feedback [--logs] <what went wrong>` | Send feedback to Codex; logs are sent only with `--logs` |

tcodex also shows model-specific rate limits next to the main one, and
offers models from Codex's own catalog with each model's efforts.

### What tcodex leaves out

These Claude features have no Codex equivalent:

- the context breakdown (the bar shows token totals instead)
- `/btw`, `/plan`, `/agents`, `/permissions`, `/usage`, `/reload`, `/cd`,
  `/mcp login` and Ctrl+B
- restoring files on rewind (Codex rewinds only the conversation)
- prompt suggestions and the thinking budget
- the plan-approval overlay (leave plan mode with Shift+Tab)

## OpenCode

`topencode` drives the `opencode` CLI and offers its own commands in
addition to the shared ones:

- `/agents` lists OpenCode's agents, `/permissions` its permission rules.
- `/btw <question>` asks a side question.
- `/diff` shows what the current or last turn changed.
- `/export` saves the conversation as JSON in this directory.
- `!command` runs a shell command in the conversation, like OpenCode does.

`topencode` passes no model by default; `-agent NAME` runs turns with a
named agent such as build or plan.

## Hermes

Hermes is a supported native Tern agent, but it is **not** served by a Go
backend in this repo. It is served by the separate
[hermes-for-tern](https://github.com/thefullctx/hermes-for-tern) frontend
(a Python TSP frontend), which also installs the `hermes` CLI wrapper. The
plugin palette's **New Hermes chat** command runs that wrapper; there is no
`thermes` binary here.
