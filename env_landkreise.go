package main

import (
	"github.com/spf13/cobra"
)

// EnvLandkreise encapsulates the environment for the CLI landkreise handler.
type EnvLandkreise struct {
}

// EnvLandkreiseFrom creates a EnvLandkreise instance from a given cobra command and its args.
func EnvLandkreiseFrom(command *cobra.Command, args []string) (EnvLandkreise, error) {
	return EnvLandkreise{}, nil
}

func (e *EnvLandkreise) String() string {
	return ""
}
