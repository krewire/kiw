package commands

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/krewire/guild"
	"github.com/krewire/libs/core"
)

// RegisterGuild registers flags for the guild command group.
func RegisterGuild(fs *flag.FlagSet) {
	fs.Bool("force", false, "overwrite existing managed files without prompting")
	fs.Bool("dry-run", false, "report the files that would be written without writing")
}

// RunGuild dispatches the guild sub-commands. Currently only "install" is
// implemented; with no sub-command the wizard starts interactively.
func RunGuild(fs *flag.FlagSet) core.ExitCode {
	sub := fs.Arg(0)
	switch sub {
	case "install":
		return runGuildInstall(fs, os.Stdin, os.Stdout)
	case "":
		return usageMessage("usage: kiw guild install [target] [--force] [--dry-run]")
	default:
		return usageMessage(fmt.Sprintf("unknown guild sub-command %q (supported: install)", sub))
	}
}

// runGuildInstall installs the Guild template into a target directory. When
// the target is not supplied it is asked for interactively; existing managed
// files prompt for confirmation unless --force is given.
func runGuildInstall(fs *flag.FlagSet, in io.Reader, out io.Writer) core.ExitCode {
	target := fs.Arg(1)
	if target == "" {
		line, err := promptLine(in, out, "Target directory (enter `.` for current): ")
		if err != nil {
			return fail(err)
		}
		target = line
	}
	target = strings.TrimSpace(target)
	if target == "" {
		target = "."
	}

	force := boolFlag(fs, "force")
	dryRun := boolFlag(fs, "dry-run")
	for _, arg := range fs.Args()[1:] {
		switch arg {
		case "--force":
			force = true
		case "--dry-run":
			dryRun = true
		}
	}

	if !force && !dryRun {
		conflicts, err := detectManagedConflicts(target)
		if err != nil {
			return fail(err)
		}
		if len(conflicts) > 0 {
			fmt.Fprintln(out, "The following managed files already exist:")
			for _, c := range conflicts {
				fmt.Fprintf(out, "  %s\n", c)
			}
			answer, err := promptLine(in, out, "Overwrite them? [y/N]: ")
			if err != nil {
				return fail(err)
			}
			if !isYes(answer) {
				fmt.Fprintln(out, "Aborted.")
				return core.ExitCodeUsage
			}
			force = true
		}
	}

	opts := []guild.Option{}
	if force {
		opts = append(opts, guild.WithForce())
	}
	if dryRun {
		opts = append(opts, guild.WithDryRun())
	}

	target, err := absPath(target)
	if err != nil {
		return fail(err)
	}
	created, err := guild.Install(target, opts...)
	if err != nil {
		return guildInstallError(err)
	}

	if dryRun {
		fmt.Fprintln(out, "Dry run — would install Guild into "+target)
	} else {
		slog.Info("installed guild template", "dir", target, "files", len(created))
		fmt.Fprintln(out, "Installed Guild into "+target)
	}
	for _, path := range created {
		fmt.Fprintln(out, "created "+path)
	}

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Next steps:")
	fmt.Fprintln(out, "  1. cd "+target)
	fmt.Fprintln(out, "  2. Open opencode there.")
	fmt.Fprintln(out, "  3. Run /kickoff so the agent maps the project.")
	return core.ExitCodeSuccess
}

// absPath returns the absolute form of target.
func absPath(target string) (string, error) {
	if strings.HasPrefix(target, "~") {
		return "", core.UsageError("shell expansion is not applied; use an absolute or relative path")
	}
	return filepath.Abs(target)
}

// detectManagedConflicts reports managed guild paths that already exist under
// target, mirroring the library's conflict detection for prompting.
func detectManagedConflicts(target string) ([]string, error) {
	var conflicts []string
	for _, rel := range guild.Managed() {
		p := filepath.Join(target, rel)
		if _, err := os.Lstat(p); err == nil {
			conflicts = append(conflicts, p)
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	return conflicts, nil
}

// guildInstallError maps the library's sentinel errors onto exit codes.
func guildInstallError(err error) core.ExitCode {
	fmt.Fprintln(os.Stderr, "kiw:", err)
	switch {
	case errors.Is(err, guild.ErrTargetMissing), errors.Is(err, guild.ErrConflicts):
		return core.ExitCodeUsage
	default:
		return core.ExitCodeFailure
	}
}

// promptLine writes question and reads one trimmed input line.
func promptLine(in io.Reader, out io.Writer, question string) (string, error) {
	fmt.Fprint(out, question)
	sc := bufio.NewScanner(in)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return "", err
		}
		return "", nil
	}
	return strings.TrimSpace(sc.Text()), nil
}

// isYes reports whether a confirmation answer is affirmative.
func isYes(answer string) bool {
	a := strings.ToLower(strings.TrimSpace(answer))
	return a == "y" || a == "yes"
}

// usageMessage prints a usage line and returns ExitCodeUsage.
func usageMessage(msg string) core.ExitCode {
	fmt.Fprintln(os.Stderr, "kiw: "+msg)
	return core.ExitCodeUsage
}

func boolFlag(fs *flag.FlagSet, name string) bool {
	if f := fs.Lookup(name); f != nil {
		return f.Value.String() == "true"
	}
	return false
}
