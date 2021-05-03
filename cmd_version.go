package main

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func getVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "version",
		Run: handleVersionCmd,
	}
	return cmd
}

func handleVersionCmd(cmd *cobra.Command, args []string) {
	log.Info().
		Str("git version", GitVersion).
		Str("git commit", GitCommit).
		Str("git date", GitDate).
		Str("git state", GitState).
		Str("git branch", GitBranch).
		Str("git remote", GitRemote).
		Msg(ToolName + ", version " + GitVersion)
}
