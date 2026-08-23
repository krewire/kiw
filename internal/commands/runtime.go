package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/krewire/kiw/internal/rvconf"
	"github.com/krewire/libs/core"
	"github.com/krewire/libs/log"
)

// debugEnabled records whether the current command resolved debug mode, so
// low-level printers can include stack traces (KWL-P8W2N KWL-DIAGV-007).
var debugEnabled bool

// runtimeEnv is the resolved local-run context shared by serve, run, and dev
// (KWN-6K41E RND-SRV-002): the module root, its configuration, and the
// effective environment/debug switches (KWL-K4T7W).
type runtimeEnv struct {
	root  string
	cfg   *rvconf.Config
	env   core.Env
	debug bool
}

// bootRuntime resolves the module root, loads krewire.yaml and .env, and
// resolves env/debug with strict precedence: flag > KIW_ENV/KIW_DEBUG >
// krewire.yaml > default. The returned code is core.ExitCodeSuccess or a
// terminal failure/usage code.
func bootRuntime(fs *flag.FlagSet) (*runtimeEnv, core.ExitCode) {
	root, err := findRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "kiw: "+err.Error())
		return nil, core.ExitCodeUsage
	}
	c, err := rvconf.Load(root)
	if err != nil {
		return nil, fail(core.WithStack(err))
	}
	if err := c.LoadDotEnv(root); err != nil {
		return nil, fail(core.WithStack(err))
	}
	env, err := c.ResolveEnv(flagValue(fs, "env"))
	if err != nil {
		return nil, usageOrFail(err)
	}
	debug := c.ResolveDebug(flagValue(fs, "debug"), flagProvided(fs, "debug"))
	debugEnabled = debug
	log.Install(env, debug)
	return &runtimeEnv{root: root, cfg: c, env: env, debug: debug}, core.ExitCodeSuccess
}

// registerRuntimeFlags registers the flags shared by every command that
// boots a project locally.
func registerRuntimeFlags(fs *flag.FlagSet) {
	fs.String("kind", "", "force project kind: app, cli, site, or book (default auto)")
	fs.String("env", "", "target environment: local, production, or testing (default: krewire.yaml, then KIW_ENV)")
	fs.Bool("debug", false, "enable debug mode (default: krewire.yaml, then KIW_DEBUG)")
}
