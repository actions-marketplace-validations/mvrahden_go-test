package main

import (
	"strings"
	"time"

	"github.com/mvrahden/go-test/internal/config"
	"github.com/mvrahden/go-test/internal/gotestrunner"
)

// Invocation carries the resolved state for a single CLI invocation.
type Invocation struct {
	Args   []string
	Config config.ProjectConfig
}

// TagArgs returns args with -tags prepended from config, if not already set.
func (inv Invocation) TagArgs() []string {
	if inv.Config.Tags == "" {
		return inv.Args
	}
	if hasFlag(inv.Args, "-tags") {
		return inv.Args
	}
	return append([]string{"-tags=" + inv.Config.Tags}, inv.Args...)
}

// DefaultArgs returns args with tags, setup-timeout, and debounce from config,
// each only if not already set via CLI flags.
func (inv Invocation) DefaultArgs() []string {
	out := inv.TagArgs()
	if inv.Config.SetupTimeout.Duration() > 0 && !hasFlag(out, "--setup-timeout") {
		out = append([]string{"--setup-timeout=" + inv.Config.SetupTimeout.Duration().String()}, out...)
	}
	if inv.Config.Debounce.Duration() > 0 && !hasFlag(out, "--debounce") {
		out = append([]string{"--debounce=" + inv.Config.Debounce.Duration().String()}, out...)
	}
	return out
}

func hasFlag(args []string, name string) bool {
	for _, arg := range args {
		if arg == name || strings.HasPrefix(arg, name+"=") {
			return true
		}
	}
	return false
}

type ExecConfig struct {
	GoTestArgs      []string
	PackagePatterns []string
	SetupTimeout    time.Duration
	Debug           bool
	CI              bool
	JSON            bool
	UpdateSnapshots bool
}

// knownSubcommands is the set of recognized subcommands.
var knownSubcommands = map[string]bool{
	"discover": true,
	"prepare":  true,
	"generate": true,
	"scaffold": true,
	"migrate":  true,
	"spec":     true,
	"watch":    true,
	"clean":    true,
	"lint":     true,
	"refactor": true,
	"version":  true,
	"help":     true,
}

// ParseSubcommand checks the first positional argument against known
// subcommands. If it matches, it is consumed and returned along with
// the remaining args. Otherwise, subcmd is empty and the full args
// slice is returned unchanged.
func ParseSubcommand(args []string) (subcmd string, remaining []string) {
	if len(args) == 0 {
		return "", nil
	}
	first := args[0]
	if knownSubcommands[first] {
		remaining = args[1:]
		if len(remaining) == 0 {
			remaining = nil
		}
		return first, remaining
	}
	return "", args
}

func SplitArgs(inArgs []string) (ownArgs, goTestArgs []string) {
	for i := 0; i < len(inArgs); i++ {
		arg := inArgs[i]
		switch {
		case arg == "--debug" || arg == "--ci" || arg == "--spec" || arg == "--update-snapshots":
			ownArgs = append(ownArgs, arg)
		case strings.HasPrefix(arg, "--min="):
			ownArgs = append(ownArgs, arg)
		case arg == "--min" && i+1 < len(inArgs):
			ownArgs = append(ownArgs, arg, inArgs[i+1])
			i++
		case strings.HasPrefix(arg, "--setup-timeout="):
			ownArgs = append(ownArgs, arg)
		case arg == "--setup-timeout" && i+1 < len(inArgs):
			ownArgs = append(ownArgs, arg, inArgs[i+1])
			i++
		case strings.HasPrefix(arg, "--debounce="):
			ownArgs = append(ownArgs, arg)
		case arg == "--debounce" && i+1 < len(inArgs):
			ownArgs = append(ownArgs, arg, inArgs[i+1])
			i++
		default:
			goTestArgs = append(goTestArgs, arg)
		}
	}
	return ownArgs, goTestArgs
}

func ExtractPackagePatterns(goTestArgs []string) []string {
	var patterns []string
	for _, arg := range goTestArgs {
		if arg == "-args" {
			break
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		if gotestrunner.LooksLikePackagePattern(arg) {
			patterns = append(patterns, arg)
		}
	}
	if len(patterns) == 0 {
		return []string{"."}
	}
	return patterns
}

func extractTagsFlag(args []string) (tags string, remaining []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if v, ok := strings.CutPrefix(arg, "-tags="); ok {
			tags = v
		} else if arg == "-tags" && i+1 < len(args) {
			tags = args[i+1]
			i++
		} else {
			remaining = append(remaining, arg)
		}
	}
	return tags, remaining
}
