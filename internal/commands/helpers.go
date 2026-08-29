package commands

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/krewire/kiw/internal/scaffold"
	"github.com/krewire/libs/core"
	"github.com/krewire/libs/vein"
)

func flagValue(fs *flag.FlagSet, name string) string {
	if f := fs.Lookup(name); f != nil {
		return f.Value.String()
	}
	return ""
}

func commandError(err error) core.ExitCode {
	fmt.Fprint(os.Stderr, vein.FormatTree(err))
	switch {
	case errors.Is(err, scaffold.ErrInvalidName),
		errors.Is(err, scaffold.ErrProjectExists),
		errors.Is(err, scaffold.ErrNotEmpty),
		errors.Is(err, scaffold.ErrConflict):
		return core.ExitCodeUsage
	default:
		return core.ExitCodeFailure
	}
}

func fail(err error) core.ExitCode {
	fmt.Fprint(os.Stderr, vein.FormatTree(err))
	printStackIfDebug(err)
	return core.ExitCodeFailure
}

// printStackIfDebug renders an attached stack after the error line when the
// command resolved debug mode (KWL-P8W2N KWL-DIAGV-007).
func printStackIfDebug(err error) {
	if !debugEnabled {
		return
	}
	frames := vein.StackOf(err)
	if len(frames) == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, "stack trace (newest first):")
	fmt.Fprint(os.Stderr, vein.FormatStack(frames))
}

// usageOrFail prints err and maps core usage errors to exit code 2
// (KWL-K4T7W KWL-ENVV-007).
func usageOrFail(err error) core.ExitCode {
	var ce interface{ ExitCode() core.ExitCode }
	if errors.As(err, &ce) && ce.ExitCode() == core.ExitCodeUsage {
		fmt.Fprint(os.Stderr, vein.FormatTree(err))
		printStackIfDebug(err)
		return core.ExitCodeUsage
	}
	return fail(err)
}

// flagProvided reports whether the named flag was explicitly set on the
// command line, as opposed to holding its zero value.
func flagProvided(fs *flag.FlagSet, name string) bool {
	set := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			set = true
		}
	})
	return set
}

// truthStr renders a boolean as the KIW_DEBUG wire value.
func truthStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// childEnviron composes the child-process environment for run and dev,
// exporting APP_ADDR plus KIW_ENV/KIW_DEBUG (KWL-K4T7W KWL-ENVV-006).
// Vein is applied: env via vein.Env.
func childEnviron(base []string, addr string, env vein.Env, debug bool) []string {
	e := append([]string(nil), base...)
	if addr != "" {
		e = append(e, "APP_ADDR="+addr)
	}
	return append(e, "KIW_ENV="+env.String(), "KIW_DEBUG="+truthStr(debug))
}
