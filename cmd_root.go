package main

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/iancoleman/strcase"
	"github.com/jojomi/covid-mailer/rki"
	"github.com/jojomi/strtpl"
	"github.com/peterbourgon/diskv"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"gopkg.in/gomail.v2"
)

func getRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: strcase.ToKebab(ToolName),
		Run: handleRootCmd,
	}

	pf := cmd.PersistentFlags()
	pf.BoolP("verbose", "v", false, "activate more verbose output")

	f := cmd.Flags()
	f.StringSliceP("landkreis", "l", []string{"LK Berlin"}, "Landkreis-Name(n)")
	err := cmd.MarkFlagRequired("landkreis")
	if err != nil {
		log.Fatal().Err(err).Msg("Flag landkreis nicht als required markierbar")
	}
	f.StringP("recipient", "r", "", "E-Mail-Empfänger")
	err = cmd.MarkFlagRequired("recipient")
	if err != nil {
		log.Fatal().Err(err).Msg("Flag recipient nicht als required markierbar")
	}
	f.BoolP("once-per-day", "o", false, "Nur einmal pro Kalendertag verschicken")
	f.BoolP("dry-run", "d", false, "Nur testen?")
	f.BoolP("use-cache", "c", false, "Lokale Dateien nutzen, wenn vorhanden?")
	f.IntP("history-days", "i", 14, "Anzahl der Tage, die in der Historie angezeigt werden sollen")

	cmd.AddCommand(getLandkreiseCmd())

	return cmd
}

func handleRootCmd(cmd *cobra.Command, args []string) {
	env, err := EnvRootFrom(cmd, args)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not build root env")
	}
	handleRoot(env)
}

var (
	//go:embed templates
	templates    embed.FS
	datetimeFull = time.RFC3339
	dateSortable = "2006-01-02"
)

func handleRoot(env EnvRoot) {
	setLoggerVerbosity(env.Verbose)
	config := mustGetConfig("config.yml")

	var (
		now        = time.Now()
		diskvStore *diskv.Diskv
		key        string
		lastStand  time.Time
		err        error
	)
	if env.OncePerDay {
		r := regexp.MustCompile(`\W`)
		key = r.ReplaceAllString("last stand "+env.String(), "_")
		flatTransform := func(s string) []string { return []string{} }
		diskvStore = diskv.New(diskv.Options{
			BasePath:     "cache-dir",
			Transform:    flatTransform,
			CacheSizeMax: 1024 * 1024,
		})
		data, err := diskvStore.Read(key)
		if err == nil {
			lastStand, err = time.Parse(datetimeFull, string(data))
			if err != nil {
				log.Fatal().Err(err).Msg("Letzter Stand ungültig")
			}
		}
		// mit Daten von heute schon gelaufen
		if lastStand.Format(dateSortable) == now.Format(dateSortable) {
			log.Info().Time("Letzter verarbeiteter RKI-Stand", lastStand).Time("now", now).Msg("Mit Daten von heute bereits gelaufen")
			os.Exit(0)
		}
	}

	if !env.UseCache {
		err = rki.DownloadInzidenzData()
		if err != nil {
			log.Fatal().Err(err).Msg("Download der Daten vom RKI fehlgeschlagen")
		}
	}

	f, err := rki.OpenFile(true)
	if err != nil {
		log.Fatal().Err(err).Msg("Öffnen der Inzidenzdaten fehlgeschlagen")
	}

	// Datenstand prüfen
	stand, err := rki.GetStand(f)
	if err != nil {
		log.Fatal().Err(err).Msg("Stand konnte nicht ermittelt werden")
	} else {
		log.Info().Time("Stand", stand).Msg("Stand ermittelt")
	}

	// Stand von gestern? -> Abbruch
	if stand.Format(dateSortable) < now.Format(dateSortable) {
		log.Info().Time("RKI-Daten", stand).Time("now", now).Msg("Datenstand beim RKI noch nicht von heute. Täglicher Upload erfolgt üblicherweise gegen 9 Uhr vormittags.")
		os.Exit(0)
	}

	// mit gleichem Stand schon gelaufen?
	if env.OncePerDay && lastStand.Format(dateSortable) == stand.Format(dateSortable) {
		log.Info().Msg("Soll nur einmal pro Tag laufen und ist bereits durchgelaufen")
		os.Exit(0)
	}

	landkreise := env.Landkreise
	landkreisdaten := make([]rki.Landkreisdaten, 0)
	for _, landkreis := range landkreise {
		row, err := rki.GetRowByName(f, landkreis)
		if err != nil {
			log.Fatal().Err(err).Msg("Landkreis nicht gefunden")
		} else {
			log.Info().Str("Name", landkreis).Int("Zeile", row).Msg("Landkreis gefunden")
		}

		inzidenzen, err := rki.GetInzidenzenByRow(f, row)
		if err != nil {
			log.Fatal().Err(err).Msg("Inzidenzermittlung fehlgeschlagen")
		} else {
			log.Info().Int("Anzahl Tage", len(inzidenzen)).Msg("7-Tage-Inzidenzen ermittelt")
		}

		landkreisdaten = append(landkreisdaten, rki.Landkreisdaten{
			Name:       landkreis,
			Inzidenzen: inzidenzen,
		})
	}

	templateName := filepath.Join("templates", "email.tpl")
	bodyTemplate, err := templates.ReadFile(templateName)
	if err != nil {
		log.Fatal().Err(err).Str("Dateiname", templateName).Msg("Template nicht lesbar")
	}
	body, err := strtpl.EvalHTMLWithFuncMap(string(bodyTemplate), template.FuncMap{
		"last":     rki.Last,
		"lastN":    rki.LastN,
		"lastNRev": rki.LastNRev,
	}, struct {
		Stand      time.Time
		Landkreise []rki.Landkreisdaten
		Env        EnvRoot
	}{
		Stand:      stand,
		Landkreise: landkreisdaten,
		Env:        env,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("E-Mail-Inhalt nicht erzeugbar")
	}

	aktuell := make([]string, len(landkreisdaten))
	for i, l := range landkreisdaten {
		aktuell[i] = fmt.Sprintf("%.1f", rki.Last(l.Inzidenzen).Value)
	}

	subject, err := strtpl.Eval(`[ {{- .Aktuell -}} ] Corona-Zahlen und 7-Tage-Inzidenz, Stand: {{ .Stand.Format "02.01.2006" }}`, map[string]interface{}{
		"Aktuell": strings.Join(aktuell, " · "),
		"Stand":   stand,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Betreff nicht erzeugbar")
	}

	if env.DryRun {
		log.Info().Str("Betreff", subject).Str("Inhalt", body).Msg("E-Mail")
	} else {
		err = email(config.SMTP, env.Recipient, subject, body)
		if err != nil {
			log.Fatal().Err(err).Msg("E-Mail-Versand fehlgeschlagen")
		} else {
			log.Info().Str("SMTP server", config.SMTP.Server).Str("from", config.SMTP.From).Msg("E-Mail-Versand erfolgreich")
		}
	}

	// save the execution if required
	if env.OncePerDay && diskvStore != nil && !env.DryRun {
		err := diskvStore.Write(key, []byte(stand.Format(datetimeFull)))
		if err != nil {
			log.Fatal().Err(err).Msg("Konnte nicht in Cache schreiben")
		}
	}
}

func email(config SMTPConfig, to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(config.Server, config.Port, config.User, config.Password)
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
