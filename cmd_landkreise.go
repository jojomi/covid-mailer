package main

import (
	"fmt"

	"github.com/jojomi/covid-mailer/rki"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func getLandkreiseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "landkreise",
		Run: handleLandkreiseCmd,
	}
	return cmd
}

func handleLandkreiseCmd(cmd *cobra.Command, args []string) {
	f, err := rki.OpenFile(true)
	if err != nil {
		log.Fatal().Err(err).Msg("Öffnen der Inzidenzdaten fehlgeschlagen")
	}

	landkreise, err := rki.GetLandkreise(f)
	if err != nil {
		log.Fatal().Err(err).Msg("Ermittlung der Landkreise fehlgeschlagen")
	}

	for _, landkreis := range landkreise {
		fmt.Println(landkreis)
	}
}
