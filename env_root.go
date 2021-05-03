package main

import (
	"strings"
	"text/template"

	"github.com/jojomi/strtpl"
	"github.com/spf13/cobra"
)

// EnvRoot encapsulates the environment for the CLI root handler.
type EnvRoot struct {
	Landkreise  []string
	Recipient   string
	HistoryDays int
	OncePerDay  bool
	UseCache    bool
	Verbose     bool
	DryRun      bool
}

// EnvRootFrom builds a EnvRoot from a given cobra command and its args.
func EnvRootFrom(command *cobra.Command, args []string) (EnvRoot, error) {
	var (
		e   = EnvRoot{}
		err error
	)
	pf := command.PersistentFlags()
	f := command.Flags()

	e.Landkreise, err = f.GetStringSlice("landkreis")
	if err != nil {
		return e, err
	}

	e.Recipient, err = f.GetString("recipient")
	if err != nil {
		return e, err
	}

	e.HistoryDays, err = f.GetInt("history-days")
	if err != nil {
		return e, err
	}

	e.OncePerDay, err = f.GetBool("once-per-day")
	if err != nil {
		return e, err
	}

	e.UseCache, err = f.GetBool("use-cache")
	if err != nil {
		return e, err
	}

	e.Verbose, err = pf.GetBool("verbose")
	if err != nil {
		return e, err
	}

	e.DryRun, err = f.GetBool("dry-run")
	if err != nil {
		return e, err
	}

	return e, nil
}

func (e *EnvRoot) String() string {
	return strtpl.MustEvalWithFuncMap(strings.TrimSpace(`
	Landkreise: {{ join .Landkreise "," }}
	Recipient: {{ .Recipient }}
	UseCache: {{ .UseCache }}
	HistoryDays: {{ .HistoryDays }}
	Verbose: {{ .Verbose }}
	DryRun: {{ .DryRun }}
	`), template.FuncMap{
		"join": strings.Join,
	}, e)
}
