package main

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// Sign to ecliptic longitude offset
var signOffset = map[string]float64{
	"Aries": 0, "Taurus": 30, "Gemini": 60, "Cancer": 90,
	"Leo": 120, "Virgo": 150, "Libra": 180, "Scorpio": 210,
	"Sagittarius": 240, "Capricorn": 270, "Aquarius": 300, "Pisces": 330,
}

// Standard aspects: name -> angle
var standardAspects = []struct {
	name  string
	angle float64
}{
	{"Conjunction", 0},
	{"Semi-Sextile", 30},
	{"Semi-Square", 45},
	{"Sextile", 60},
	{"Square", 90},
	{"Trine", 120},
	{"Sesquiquadrate", 135},
	{"Quincunx", 150},
	{"Opposition", 180},
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fixcsv <csv-file> [csv-file2 ...]")
		os.Exit(1)
	}

	for _, path := range os.Args[1:] {
		if err := fixCSV(path); err != nil {
			fmt.Printf("ERROR fixing %s: %v\n", path, err)
		}
	}
}

func fixCSV(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	f.Close()
	if err != nil {
		return err
	}

	if len(records) < 2 {
		return fmt.Errorf("no data rows")
	}

	// Find column indices from header
	header := records[0]
	colIdx := map[string]int{}
	for i, h := range header {
		// Strip BOM if present
		h = strings.TrimPrefix(h, "\ufeff")
		colIdx[h] = i
	}

	aspectCol, ok1 := colIdx["Aspect"]
	pos1DegCol, ok2 := colIdx["Pos1_Deg"]
	pos1SignCol, ok3 := colIdx["Pos1_Sign"]
	pos2DegCol, ok4 := colIdx["Pos2_Deg"]
	pos2SignCol, ok5 := colIdx["Pos2_Sign"]

	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
		return fmt.Errorf("missing required columns")
	}

	fixed := 0
	total := 0
	unchanged := 0

	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) <= pos2SignCol {
			continue
		}
		total++

		deg1, err1 := strconv.ParseFloat(row[pos1DegCol], 64)
		deg2, err2 := strconv.ParseFloat(row[pos2DegCol], 64)
		if err1 != nil || err2 != nil {
			continue
		}

		off1, ok1 := signOffset[row[pos1SignCol]]
		off2, ok2 := signOffset[row[pos2SignCol]]
		if !ok1 || !ok2 {
			fmt.Printf("  Row %d: unknown sign %q or %q\n", i+1, row[pos1SignCol], row[pos2SignCol])
			continue
		}

		lon1 := off1 + deg1
		lon2 := off2 + deg2

		// Compute angular distance (0-180)
		diff := math.Abs(lon1 - lon2)
		if diff > 180 {
			diff = 360 - diff
		}

		// Find closest standard aspect
		correctAspect := findClosestAspect(diff)
		oldAspect := row[aspectCol]

		if correctAspect != oldAspect {
			row[aspectCol] = correctAspect
			fixed++
			if fixed <= 10 {
				fmt.Printf("  Row %d: %s -> %s (distance=%.3f)\n", i+1, oldAspect, correctAspect, diff)
			}
		} else {
			unchanged++
		}
	}

	fmt.Printf("%s: %d rows, %d fixed, %d unchanged\n", path, total, fixed, unchanged)

	// Write back
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write BOM for Excel compatibility
	out.Write([]byte("\xef\xbb\xbf"))

	writer := csv.NewWriter(out)
	// Clean BOM from header before writing (we already wrote it)
	cleanHeader := make([]string, len(header))
	copy(cleanHeader, header)
	cleanHeader[0] = strings.TrimPrefix(cleanHeader[0], "\ufeff")
	records[0] = cleanHeader

	err = writer.WriteAll(records)
	if err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}

func findClosestAspect(angularDist float64) string {
	bestName := "Unknown"
	bestDiff := 999.0

	for _, a := range standardAspects {
		d := math.Abs(angularDist - a.angle)
		if d < bestDiff {
			bestDiff = d
			bestName = a.name
		}
	}

	// Sanity check: if best match is more than 5 degrees off, something is wrong
	if bestDiff > 5 {
		return fmt.Sprintf("Unknown(%.1f)", angularDist)
	}

	return bestName
}
