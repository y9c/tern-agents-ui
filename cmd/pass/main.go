// Command pass is a thin passthrough shim for an agent CLI: it finds the real
// tool on the system and execs it with every argument unchanged, so a wrapper
// named like the original always uses the latest installed version and behaves
// exactly like it. Add it to PATH ahead of the real CLI (or as `opencode`,
// `codex`, `claude`, `pi`) and it becomes a transparent drop-in.
//
// It names itself from os.Args[0], or from the TOOL_REAL environment variable
// (for example OPENCODE_REAL=/path/to/real/opencode) to force a specific
// binary. "pass" is only the source name; build it as each tool.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	tool := os.Getenv("PASS_TOOL")
	if tool == "" {
		tool = strings.TrimSuffix(filepath.Base(os.Args[0]), filepath.Ext(os.Args[0]))
	}
	if tool == "" || tool == "pass" || tool == "pass.exe" {
		tool = os.Getenv("PASS_TOOL")
	}

	real := findReal(tool)
	if real == "" {
		fmt.Fprintf(os.Stderr, "%s: could not find the real %q on PATH; install it or set %s_REAL\n", tool, tool, strings.ToUpper(tool))
		os.Exit(127)
	}

	cmd := exec.Command(real, os.Args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	// The child runs in the same session, so terminal signals (Ctrl+C) reach
	// both; on exit the shim reports the child's status.
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// findReal returns the path of the real tool, excluding the shim itself, so a
// same-named binary never execs itself. TOOL_REAL wins; otherwise it searches
// PATH for the first executable named tool that is not this shim.
func findReal(tool string) string {
	return findRealPath(tool, realPath(os.Args[0]))
}

func findRealPath(tool, self string) string {
	if r := os.Getenv(strings.ToUpper(tool) + "_REAL"); r != "" {
		return r
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" {
			dir = "."
		}
		candidate := filepath.Join(dir, tool)
		if realPath(candidate) == self {
			continue
		}
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() && fi.Mode()&0o111 != 0 {
			return candidate
		}
	}
	return ""
}

// realPath resolves symlinks so a symlink to the shim is not mistaken for the
// real tool.
func realPath(p string) string {
	if r, err := filepath.EvalSymlinks(p); err == nil {
		if a, err := filepath.Abs(r); err == nil {
			return a
		}
		return r
	}
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return p
}
