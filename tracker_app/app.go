package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Record
type Record struct {
	duration  int // duration in seconds
	exe_path  string
	url       string
	name      string
	date_id   int
	date_info DateInfo
	category  string
}

// DateInfo holds enriched information about a date
type DateInfo struct {
	DayOfWeek       string
	MonthName       string
	WeekOfYear      int
	IsMarketHoliday bool
	IsWeekend       bool
}

// App struct
type App struct {
	ctx                context.Context
	records            []Record            // loaded records
	categories         map[string]string   // map of url/exe_path to category
	reverse_categories map[string][]string // map of category to list of url/exe_paths for easy lookup

}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx
	a.records = []Record{}
	a.categories = make(map[string]string)
	a.reverse_categories = make(map[string][]string)

	// populate categories
	a.populate_categories()
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {

}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {

}

/*

Desired Functionality

- URL -> website base url
- Categorize function: map app/webstite to category (e.g., Social Media, Work, Entertainment). Pulls from a predefined list or config file.
- time_on_date function: reads the CSV file for a given date and returns the time spent (in minutes) with each application focused.
- time_per_category
*/

func (a *App) populate_categories() error {
	data_dir := filepath.Join(os.Getenv("LOCALAPPDATA"), "tracker_data")
	data_file_path := filepath.Join(data_dir, "categories.json")
	f, err := os.OpenFile(data_file_path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// by default, url/exe_path -> category is empty
	categories := map[string]string{}

	// by default, category -> url/exe_paths is empty. but pre-defined 5 categories
	reverse_categories := map[string][]string{
		"Other":         {},
		"Work":          {},
		"Productivity":  {},
		"Entertainment": {},
		"Games":         {},
	}

	a.categories = categories
	a.reverse_categories = reverse_categories

	return nil
}

// populate_records takes in the raw data from the CSV, header and all, and populates the App's records slice.
func (a *App) populate_records(records [][]string) error {
	if len(records) < 3 {
		return nil // do nothing if not enough data
	}

	// iterate through consecutive pairs of readings
	for i := 0; i < len(records)-1; i++ {
		exePath := records[i][0]
		name := records[i][1]
		url := records[i][2]

		if exePath == "Off" { // app was off or asleep: don't count the time from this to next
			continue
		}
		currentTime, err := time.Parse(time.RFC3339, records[i][1])
		if err != nil {
			continue
		}
		nextTime, err := time.Parse(time.RFC3339, records[i+1][1])
		if err != nil {
			continue
		}
		date_id := currentTime.Year()*10000 + int(currentTime.Month())*100 + currentTime.Day()

		for duration := nextTime.Sub(currentTime); duration.Seconds() <= 15; { // if longer than 15 seconds, likely computer was off or asleep
			// if this code block is entered, this is a valid record
			record := Record{
				duration: int(duration.Seconds()),

				exe_path: exePath,
				url:      url,
				name:     name,

				date_id:   date_id,
				date_info: a.enrich_date(date_id),
				category:  a.categorize(exePath, url),
			}
			a.records = append(a.records, record)
		}
	}
	return nil
}

// populate_date reads the CSV file for the given date and populates the records on that date.
func (a *App) populate_date(date int) error {
	data_dir := filepath.Join(os.Getenv("LOCALAPPDATA"), "tracker_data")
	data_file_path := filepath.Join(data_dir, fmt.Sprintf("%d.csv", date))

	f, err := os.Open(data_file_path)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	a.populate_records(records)
	return nil
}

// categorize takes in an application name and URL, and returns the category it belongs to, according to the predefined categories.
// Categorize by url if given, otherwise by exe name.
// The name is the website name if applicable, otherwise the exe name. The url is self explanatory.
func (a *App) categorize(exePath string, url string) string {
	if url != "" { // categorize by url
		for key, category := range a.categories {
			if strings.Contains(url, key) {
				return category
			}
		}
		return "Other"
	} else { // categorize by exe path
		if category, exists := a.categories[exePath]; exists {
			return category
		}
	}
	return "Other"
}

// enrich_date takes a date_id (YYYYMMDD format) and returns enriched date information
// including day of week, month name, week of year, market holiday status, and weekend status.
func (a *App) enrich_date(date_id int) DateInfo {
	// Extract year, month, and day from date_id
	year := date_id / 10000
	month := (date_id / 100) % 100
	day := date_id % 100

	// Create time.Time object
	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	// Get day of week
	weekday := t.Weekday()
	dayOfWeek := weekday.String()

	// Get month name
	monthName := t.Month().String()

	// Get ISO week of year
	_, weekOfYear := t.ISOWeek()

	// Check if weekend
	isWeekend := weekday == time.Saturday || weekday == time.Sunday

	// Market holiday detection - placeholder for future implementation
	isMarketHoliday := false

	return DateInfo{
		DayOfWeek:       dayOfWeek,
		MonthName:       monthName,
		WeekOfYear:      weekOfYear,
		IsMarketHoliday: isMarketHoliday,
		IsWeekend:       isWeekend,
	}
}
