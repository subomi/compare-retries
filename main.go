package main

import (
	"os"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/sethvargo/go-retry"
)

type strategy struct {
	name  string
	retry retry.Backoff
}

var (
	// flags
	attempts = 20

	// table sub headings
	subheadingColumns = []string{"Duration", "Cumulative Duration"}

	// algorithms
	strategies = []strategy{
		strategy{
			name:  "Linear Retry",
			retry: retry.NewConstant(1 * time.Hour),
		},
		strategy{
			name:  "Exponential Backoff",
			retry: retry.NewExponential(20 * time.Second),
		},
		strategy{
			name: "Capped Exponential Backoff",
			retry: func() retry.Backoff {
				ex := retry.NewExponential(20 * time.Second)
				return retry.WithCappedDuration(7200*time.Second, ex)
			}(),
		},
	}
)

func main() {

	// Create a new table writer using os.Stdout
	table := tablewriter.NewWriter(os.Stdout)

	// Set heading
	table.SetHeader(generateHeading([]string{""}, getNames(strategies)))

	// Set subheader
	table.Append(generateSubHeading([]string{"Attempts"}, len(strategies)))

	// Set table settings
	table.SetRowLine(true)

	var rows [][]string
	for attempt := 0; attempt < attempts; attempt++ {
		row := []string{strconv.Itoa(attempt)}

		cells := generateCells([]string{}, rows, attempt, 0)
		row = append(row, cells...)
		rows = append(rows, row)
	}

	table.AppendBulk(rows)
	table.Render()
}

func generateHeading(heading []string, strategies []string) []string {
	if len(strategies) == 0 {
		return heading
	}

	heading = append(heading, strategies[0], "")
	return generateHeading(heading, strategies[1:])
}

func generateSubHeading(subheading []string, count int) []string {
	if count == 0 {
		return subheading
	}

	subheading = append(subheading, subheadingColumns...)
	return generateSubHeading(subheading, count-1)
}

func generateCells(cells []string, rows [][]string, attempt int, idx int) []string {
	if len(strategies) == idx {
		return cells
	}

	var duration, cumulativeDuration time.Duration
	duration, _ = strategies[idx].retry.Next()
	if attempt == 0 {
		cumulativeDuration = duration
	} else {
		cell := rows[attempt-1][2*(1+idx)]
		previousCumulativeDuration, err := time.ParseDuration(cell)
		if err != nil {
			panic(err)
		}

		cumulativeDuration = duration + previousCumulativeDuration
	}

	cells = append(cells, duration.String(), cumulativeDuration.String())

	return generateCells(cells, rows, attempt, idx+1)
}

func getNames(items []strategy) []string {
	names := make([]string, len(items))
	for i, item := range items {
		names[i] = item.name
	}
	return names
}
