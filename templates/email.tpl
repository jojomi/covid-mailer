{{ range $i, $l := .Landkreise }}
    {{ with $l }}
        {{ if ne $i 0 -}}
            <hr>
        {{ end -}}
        {{- $curr := (last .Inzidenzen).Value -}}
        <p style="font-size:150%">7-Tage-Inzidenz pro 100.000 Einwohner im {{ .Name }}: <b>{{ printf "%.1f" $curr }}</b></p>

        <p>Entwicklung über die vergangenen {{ $.Env.HistoryDays }} Tage:</p>
        {{ range (lastN .Inzidenzen $.Env.HistoryDays) -}}
            <div><span style="display:inline-block;width:16ch">{{ .Date.Format "Mon, 02.01.2006" }}:</span> <span style="display:inline-block;width:6ch;text-align:right;font-weight:bold">{{ printf "%.1f" .Value }}</span></div>
        {{- end }}
    {{- end }}
{{- end }}

<p style="font-size:80%">Generiert aus den offiziellen Daten des RKI.</p>

<p><i style="font-size:70%">Stand: {{ .Stand.Format "02.01.2006 15:04" }} Uhr</i></p>