# covid-mailer

Dieses Programm ermöglicht es, die tagesaktuellen RKI-Zahlen der 7-Tage-Corona-Inzidenzen per E-Mail zu verschicken.

## Vorbereitung

Konfiguration für ein E-Mail-Konto zum Versand der Emails muss in der Datei `config.yml` hinterlegt werden, siehe [`config.yml.example`](config.yml.example).

## Parameter

``` bash
covid-mailer --help
Usage:
  covid-mailer [flags]
  covid-mailer [command]

Available Commands:
  help        Help about any command
  landkreise  
  version     

Flags:
  -d, --dry-run             Nur testen?
  -h, --help                help for covid-mailer
  -l, --landkreis strings   Landkreis-Name(n)
  -o, --once-per-day        Nur einmal pro Kalendertag verschicken
  -r, --recipient string    E-Mail-Empfänger
  -c, --use-cache           Lokale Dateien nutzen, wenn vorhanden?
  -v, --verbose             activate more verbose output

Use "covid-mailer [command] --help" for more information about a command.
```

## Verwendung

In der crontab kann es so aussehen (für mehrere Empfänger mit verschiedenen Landkreisen):

```
*/5 6-13 * * * cd /src/github.com/jojomi/covid-mailer && covid-mailer -r whatever@gmail.com -l "LK München" -o
*/5 6-13 * * * cd /src/github.com/jojomi/covid-mailer && covid-mailer -r someoneelse@gmail.com -l "SK Kiel,SK München" -o -c
```