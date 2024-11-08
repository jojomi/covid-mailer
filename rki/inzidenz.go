package rki

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xuri/excelize/v2"
)

var (
	sourceURL     = "https://www.rki.de/DE/Content/InfAZ/N/Neuartiges_Coronavirus/Daten/Fallzahlen_Inzidenz_aktualisiert.xlsx?__blob=publicationFile"
	targetFile    = "./data/rki-data-daily.xlsx"
	sheetInzidenz = "LK_7-Tage-Inzidenz (fixiert)"
)

func DownloadInzidenzData() error {
	return downloadFile(targetFile, sourceURL)
}

func OpenFile(allowCache bool) (*excelize.File, error) {
	if _, err := os.Stat("file-exists2.file"); os.IsNotExist(err) && allowCache {
		err = DownloadInzidenzData()
		if err != nil {
			return nil, err
		}
	}
	return excelize.OpenFile(targetFile)
}

type SiebenTageInzidenz struct {
	Date  time.Time
	Value float64
}

type Landkreisdaten struct {
	Name       string
	Inzidenzen []SiebenTageInzidenz
}

func GetRowByName(f *excelize.File, search string) (int, error) {
	var (
		value string
		err   error
	)
	for i := 6; ; i++ {
		// get cell and search
		value, err = f.GetCellValue(sheetInzidenz, "B"+strconv.Itoa(i))
		if err != nil {
			log.Fatal().Err(err).Msg("Fehler bei der Landkreissuche")
		}

		if value == search {
			return i, nil
		}

		// empty? done with error
		if value == "" {
			break
		}
	}
	return 0, fmt.Errorf("Landkreis %s nicht gefunden", search)
}

func GetStand(f *excelize.File) (time.Time, error) {
	value, err := f.GetCellValue(sheetInzidenz, "A2")
	if err != nil {
		return time.Time{}, err
	}
	t, err := time.Parse("Stand: 02.01.2006 15:04:05", value)
	if err != nil {
		return getStandNew(f)
	}
	return t, nil
}

func getStandNew(f *excelize.File) (time.Time, error) {

	var (
		value  string
		err    error
		last   string
		column = 4
	)

	for {
		value, err = f.GetCellValue(sheetInzidenz, GetColumnName(column)+"5")
		if err != nil {
			return time.Time{}, err
		}
		if value == "" {
			break
		}
		last = value
		column++
	}

	t, err := time.Parse("01-02-06", last)
	if err != nil {
		return time.Time{}, err
	}

	return t, nil
}

func GetLandkreise(f *excelize.File) ([]string, error) {
	result := make([]string, 0, 200)
	var (
		row   = 6
		value string
		err   error
	)
	for {
		value, err = f.GetCellValue(sheetInzidenz, "B"+strconv.Itoa(row))
		if err != nil {
			return result, err
		}
		if value == "" {
			break
		}

		result = append(result, value)
		row++
	}
	return result, nil
}

func GetInzidenzenByRow(f *excelize.File, row int) ([]SiebenTageInzidenz, error) {
	var (
		value string
		err   error

		dateRowStr = "5"
		rowStr     = strconv.Itoa(row)
	)

	result := make([]SiebenTageInzidenz, 0, 100)

	for i := 4; ; i++ {
		value, err = f.GetCellValue(sheetInzidenz, GetColumnName(i)+rowStr)
		if err != nil {
			log.Fatal().Err(err).Str("Zelle", GetColumnName(i)+rowStr).Msg("Konnte Inzidenz nicht lesen")
		}
		if value == "" {
			break
		}
		inzidenzFloat, err := strconv.ParseFloat(value, 64)
		if err != nil {
			log.Fatal().Err(err).Str("Zelle", GetColumnName(i)+rowStr).Str("Wert", value).Msg("Inzidenz-Wert ungültig (keine Zahl)")
		}

		value, err = f.GetCellValue(sheetInzidenz, GetColumnName(i)+dateRowStr)
		if err != nil {
			log.Fatal().Err(err).Str("Zelle", GetColumnName(i)+rowStr).Msg("Konnte Inzidenz nicht lesen")
		}
		date, err := time.Parse("02.01.2006", value)
		if err != nil {
			date, err = time.Parse("01-02-06", value)
			if err != nil {
				log.Fatal().Err(err).Str("Zelle", GetColumnName(i)+rowStr).Str("Rohwert", value).Msg("Datumsformat ungültig")
			}
		}

		result = append(result, SiebenTageInzidenz{
			Date:  date,
			Value: inzidenzFloat,
		})
	}
	return result, nil
}

func GetColumnName(columnNumber int) string {
	result := ""

	var (
		rem int
	)
	for columnNumber > 0 {
		// Find remainder
		rem = columnNumber % 26

		// If remainder is 0, add 'Z'
		if rem == 0 {
			result = "Z" + result
			columnNumber = (columnNumber / 26) - 1
		} else {
			// if remainder is non-zero
			result = string(rune((rem-1)+'A')) + result
			columnNumber = columnNumber / 26
		}
	}

	return result
}

func downloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func Last(a []SiebenTageInzidenz) SiebenTageInzidenz {
	return a[len(a)-1]
}

func LastN(a []SiebenTageInzidenz, count int) []SiebenTageInzidenz {
	result := make([]SiebenTageInzidenz, 0, count)
	for i := int(math.Max(0.0, float64(len(a)-count))); i < len(a); i++ {
		result = append(result, a[i])
	}
	return result
}

func LastNRev(a []SiebenTageInzidenz, count int) []SiebenTageInzidenz {
	result := make([]SiebenTageInzidenz, 0, count)
	for i := len(a) - 1; i >= int(math.Max(0.0, float64(len(a)))); i-- {
		result = append(result, a[i])
	}
	return result
}
