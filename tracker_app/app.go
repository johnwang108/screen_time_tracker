package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/url"
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

// Grouper represents a dimension to group records by
type Grouper string

const (
	GroupByDate      Grouper = "date"
	GroupByWeek      Grouper = "week"
	GroupByMonth     Grouper = "month"
	GroupByYear      Grouper = "year"
	GroupByDayOfWeek Grouper = "day_of_week"
	GroupByIsWeekend Grouper = "is_weekend"
	GroupByCategory  Grouper = "category"
	GroupByURL       Grouper = "url"
	GroupByExePath   Grouper = "exe_path"
	GroupByName      Grouper = "name"
)

// CategoryItems holds the sites and apps assigned to a category
type CategoryItems struct {
	Sites []string `json:"sites"`
	Apps  []string `json:"apps"`
}

// Aggregation represents aggregated time across multiple records
type Aggregation struct {
	Groupers map[string]interface{} `json:"groupers"`
	Duration int                    `json:"duration"` // total duration in seconds
}

// CategoriesResponse is the shape returned to the frontend
type CategoriesResponse struct {
	Categories map[string]CategoryItems `json:"categories"`
	Order      []string                 `json:"order"`
}

// App struct
type App struct {
	ctx                  context.Context
	records              []Record                 // loaded records
	categories           map[string]string        // map of url/exe_path to category
	reverse_categories   map[string]CategoryItems // map of category to its sites and apps
	category_order       []string                 // display order of categories
	url_truncation_rules map[string][]string      // map of base domain to list of truncation patterns
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
	a.reverse_categories = make(map[string]CategoryItems)
	a.url_truncation_rules = make(map[string][]string)

	// populate categories
	a.populate_categories()
	// load URL truncation rules
	a.loadURLTruncationRules()
}

// domReady is called after front-end resources have been loaded
func (a *App) domReady(ctx context.Context) {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Now().Location())
	end := time.Now()

	var dates []int // list of date_ids since 20200101

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		date_id := d.Year()*10000 + int(d.Month())*100 + d.Day()
		dates = append(dates, date_id)
	}

	for _, date_id := range dates {
		a.populate_date(date_id)
	}
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
	data_file_path := filepath.Join(data_dir, "preferences.json")

	// Try to read existing preferences
	file, err := os.Open(data_file_path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	var rawConfig map[string]json.RawMessage
	if err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		if decErr := decoder.Decode(&rawConfig); decErr != nil {
			rawConfig = nil
		}
	}

	// Try to parse categories and order from config
	reverseCategories := map[string]CategoryItems{}
	categories := map[string]string{}
	var categoryOrder []string

	if rawConfig != nil {
		if catRaw, exists := rawConfig["categories"]; exists {
			if jsonErr := json.Unmarshal(catRaw, &reverseCategories); jsonErr != nil {
				reverseCategories = map[string]CategoryItems{}
			}
		}
		if orderRaw, exists := rawConfig["category_order"]; exists {
			json.Unmarshal(orderRaw, &categoryOrder)
		}
	}

	// If no categories found, initialize defaults and save
	if len(reverseCategories) == 0 {
		defaultOrder := []string{"Work", "Productivity", "Entertainment", "Games"}
		reverseCategories = map[string]CategoryItems{
			"Work":          {Sites: []string{}, Apps: []string{}},
			"Productivity":  {Sites: []string{}, Apps: []string{}},
			"Entertainment": {Sites: []string{}, Apps: []string{}},
			"Games":         {Sites: []string{}, Apps: []string{}},
		}
		a.reverse_categories = reverseCategories
		a.categories = categories
		a.category_order = defaultOrder
		a.saveCategories()
		return nil
	}

	// Ensure category_order is in sync with the categories map
	if len(categoryOrder) == 0 {
		for name := range reverseCategories {
			categoryOrder = append(categoryOrder, name)
		}
	}

	// Build forward map (identifier -> category) from reverse map
	for categoryName, items := range reverseCategories {
		for _, site := range items.Sites {
			categories[site] = categoryName
		}
		for _, app := range items.Apps {
			categories[app] = categoryName
		}
	}

	a.categories = categories
	a.reverse_categories = reverseCategories
	a.category_order = categoryOrder
	return nil
}

// saveCategories persists the current reverse_categories to preferences.json,
// preserving other keys like url_truncation
func (a *App) saveCategories() error {
	data_dir := filepath.Join(os.Getenv("LOCALAPPDATA"), "tracker_data")
	data_file_path := filepath.Join(data_dir, "preferences.json")

	// Read existing file to preserve other keys
	rawConfig := map[string]json.RawMessage{}
	if file, err := os.Open(data_file_path); err == nil {
		decoder := json.NewDecoder(file)
		decoder.Decode(&rawConfig)
		file.Close()
	}

	// Marshal categories and order, then update keys
	catBytes, err := json.Marshal(a.reverse_categories)
	if err != nil {
		return err
	}
	rawConfig["categories"] = catBytes

	orderBytes, err := json.Marshal(a.category_order)
	if err != nil {
		return err
	}
	rawConfig["category_order"] = orderBytes

	// Write back with indentation
	output, err := json.MarshalIndent(rawConfig, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(data_file_path, output, 0644)
}

// GetCategories returns the categories and their display order for the frontend
func (a *App) GetCategories() CategoriesResponse {
	return CategoriesResponse{
		Categories: a.reverse_categories,
		Order:      a.category_order,
	}
}

// SetItemCategory moves an identifier to a new category (or uncategorizes if category is "")
func (a *App) SetItemCategory(identifier string, category string, isApp bool) error {
	// Remove from old category if it exists
	if oldCategory, exists := a.categories[identifier]; exists {
		items := a.reverse_categories[oldCategory]
		if isApp {
			items.Apps = removeFromSlice(items.Apps, identifier)
		} else {
			items.Sites = removeFromSlice(items.Sites, identifier)
		}
		a.reverse_categories[oldCategory] = items
		delete(a.categories, identifier)
	}

	// Add to new category (unless uncategorizing)
	if category != "" {
		items := a.reverse_categories[category]
		if isApp {
			items.Apps = append(items.Apps, identifier)
		} else {
			items.Sites = append(items.Sites, identifier)
		}
		a.reverse_categories[category] = items
		a.categories[identifier] = category
	}

	if err := a.saveCategories(); err != nil {
		return err
	}
	a.recategorizeRecords()
	return nil
}

// CreateCategory adds a new empty category and appends it to the display order
func (a *App) CreateCategory(name string) error {
	if _, exists := a.reverse_categories[name]; exists {
		return fmt.Errorf("category '%s' already exists", name)
	}
	a.reverse_categories[name] = CategoryItems{Sites: []string{}, Apps: []string{}}
	a.category_order = append(a.category_order, name)
	return a.saveCategories()
}

// ReorderCategories updates the display order of categories
func (a *App) ReorderCategories(order []string) error {
	// Validate that the order contains exactly the existing categories
	if len(order) != len(a.reverse_categories) {
		return fmt.Errorf("order length does not match number of categories")
	}
	for _, name := range order {
		if _, exists := a.reverse_categories[name]; !exists {
			return fmt.Errorf("unknown category '%s'", name)
		}
	}
	a.category_order = order
	return a.saveCategories()
}

// recategorizeRecords re-applies categorization to all loaded records
func (a *App) recategorizeRecords() {
	for i := range a.records {
		a.records[i].category = a.categorize(a.records[i].exe_path, a.records[i].url)
	}
}

// removeFromSlice removes the first occurrence of item from slice
func removeFromSlice(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// loadURLTruncationRules loads URL truncation patterns from preferences.json
func (a *App) loadURLTruncationRules() error {
	data_dir := filepath.Join(os.Getenv("LOCALAPPDATA"), "tracker_data")
	data_file_path := filepath.Join(data_dir, "preferences.json")

	file, err := os.Open(data_file_path)
	if err != nil {
		// If file doesn't exist, just use empty rules
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var config struct {
		URLTruncation map[string][]string `json:"url_truncation"`
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		// If JSON is malformed or empty, just use empty rules
		return nil
	}

	a.url_truncation_rules = config.URLTruncation

	return nil
}

// matchURLPattern attempts to match a URL path against a pattern with wildcards
// Returns the truncated URL and whether it matched
func (a *App) matchURLPattern(fullURL string, baseDomain string, pattern string) (string, bool) {
	// Parse the full URL to extract the path
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return "", false
	}

	path := parsedURL.Path
	if path == "" {
		path = "/"
	}

	// Split pattern and path into segments
	patternSegments := strings.Split(strings.Trim(pattern, "/"), "/")
	pathSegments := strings.Split(strings.Trim(path, "/"), "/")

	// Handle empty path case
	if len(pathSegments) == 1 && pathSegments[0] == "" {
		pathSegments = []string{}
	}

	// Match each segment
	matchedSegments := []string{}
	for i, patternSeg := range patternSegments {
		if i >= len(pathSegments) {
			return "", false // Not enough path segments
		}

		if patternSeg == "*" {
			// Wildcard matches any single segment
			matchedSegments = append(matchedSegments, pathSegments[i])
		} else if patternSeg == pathSegments[i] {
			// Exact match
			matchedSegments = append(matchedSegments, pathSegments[i])
		} else {
			return "", false // No match
		}
	}

	// Build truncated URL
	if len(matchedSegments) > 0 {
		return baseDomain + "/" + strings.Join(matchedSegments, "/"), true
	}
	return baseDomain, true
}

// truncateURL truncates a URL to its base form or applies domain-specific patterns
func (a *App) truncateURL(rawURL string) string {
	// Handle empty URLs
	if rawURL == "" {
		return ""
	}

	// Ensure URL has a scheme for proper parsing
	parseURL := rawURL
	if !strings.Contains(rawURL, "://") {
		parseURL = "http://" + rawURL
	}

	parsedURL, err := url.Parse(parseURL)
	if err != nil {
		return rawURL // Return original if can't parse
	}

	// Extract host and remove www. prefix
	host := parsedURL.Host
	if strings.HasPrefix(host, "www.") {
		host = strings.TrimPrefix(host, "www.")
	}

	// Check if this domain has truncation patterns
	if patterns, exists := a.url_truncation_rules[host]; exists {
		// Try to match against each pattern
		for _, pattern := range patterns {
			if truncated, matched := a.matchURLPattern(parseURL, host, pattern); matched {
				return truncated
			}
		}
	}

	// No pattern matched or no patterns defined - return base domain
	return host
}

// populate_records takes in the raw data from the CSV, header and all, and populates the App's records slice.
// CSV format: name, timestamp, tabName, tabUrl, hadActivity
// Consolidates consecutive identical records and discards periods with 2+ minutes of continuous inactivity
func (a *App) populate_records(records [][]string) error {
	if len(records) < 2 {
		return nil // need at least header + 1 data row
	}

	// Track accumulated record state
	var currentExePath, currentUrl, currentName string
	var currentDateId int
	var accumulatedDuration int = 0
	var inactiveStreak int = 0

	// Helper to check if two records represent the same activity
	isSameActivity := func(exePath1, url1, name1, exePath2, url2, name2 string) bool {
		return exePath1 == exePath2 && url1 == url2 && name1 == name2
	}

	// Helper to flush accumulated record
	flushRecord := func() {
		if accumulatedDuration > 0 {
			record := Record{
				duration:  accumulatedDuration,
				exe_path:  currentExePath,
				url:       currentUrl,
				name:      currentName,
				date_id:   currentDateId,
				date_info: a.enrich_date(currentDateId),
				category:  a.categorize(currentExePath, currentUrl),
			}
			a.records = append(a.records, record)
		}
		// Reset all state
		currentExePath = ""
		currentUrl = ""
		currentName = ""
		currentDateId = 0
		accumulatedDuration = 0
		inactiveStreak = 0
	}

	// Iterate through consecutive pairs of readings (skip header row at index 0)
	for i := 1; i < len(records)-1; i++ {
		if len(records[i]) < 5 {
			continue // skip malformed rows
		}

		exePath := records[i][0]
		timestamp := records[i][1]
		tabName := records[i][2]
		tabUrl := records[i][3]
		hadActivityStr := records[i][4]

		// Skip if app was off or asleep
		if exePath == "Off" {
			flushRecord()
			continue
		}

		// Parse timestamps
		currentTime, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			continue
		}
		nextTime, err := time.Parse(time.RFC3339, records[i+1][1])
		if err != nil {
			continue
		}

		// Calculate duration
		duration := int(nextTime.Sub(currentTime).Seconds())
		if duration > 15 {
			// Likely computer was off or asleep
			flushRecord()
			continue
		}

		// Parse hadActivity
		hadActivity := hadActivityStr == "true"
		date_id := currentTime.Year()*10000 + int(currentTime.Month())*100 + currentTime.Day()

		if !hadActivity {
			// Inactive period
			if currentExePath != "" && isSameActivity(currentExePath, currentUrl, currentName, exePath, tabUrl, tabName) {
				// Same activity - add to inactive streak
				inactiveStreak += duration
				// If inactive streak reaches 2 minutes, flush and reset (don't include the inactive time)
				if inactiveStreak >= 120 {
					flushRecord()
				}
			} else {
				// Different activity or not started - flush previous if exists, don't start new
				if currentExePath != "" {
					flushRecord()
				}
				// Don't count this inactive time for the different/new activity
			}
		} else {
			// Active period
			if currentExePath == "" {
				// Start new record
				currentExePath = exePath
				currentUrl = a.truncateURL(tabUrl)
				currentName = tabName
				currentDateId = date_id
				accumulatedDuration = duration
				inactiveStreak = 0
			} else if isSameActivity(currentExePath, currentUrl, currentName, exePath, tabUrl, tabName) {
				// Same activity - add inactive streak (if < 2 min) + current duration
				if inactiveStreak < 120 {
					accumulatedDuration += inactiveStreak + duration
				} else {
					// Inactive streak was >= 2 min, so we already flushed
					// This shouldn't happen due to the flush above, but handle it
					accumulatedDuration = duration
				}
				inactiveStreak = 0
			} else {
				// Different activity - flush previous and start new
				flushRecord()
				currentExePath = exePath
				currentUrl = a.truncateURL(tabUrl)
				currentName = tabName
				currentDateId = date_id
				accumulatedDuration = duration
				inactiveStreak = 0
			}
		}
	}
	// Flush any remaining accumulated record
	flushRecord()
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

// GetAggregations aggregates records based on specified groupers and filters
func (a *App) GetAggregations(grouperNames []string, filters map[string]string) []Aggregation {
	// Map to accumulate durations: aggregation key -> duration
	aggregationMap := make(map[string]int)
	// Map to store grouper values: aggregation key -> grouper values
	grouperValuesMap := make(map[string]map[string]interface{})

	// Process each record
	for _, record := range a.records {
		// Check if record matches filters
		if !a.matchesFilters(record, filters) {
			continue
		}

		// Extract grouper values for this record
		grouperValues := make(map[string]interface{})
		keyParts := []string{}

		for _, grouperName := range grouperNames {
			value := a.extractGrouperValue(record, Grouper(grouperName))
			grouperValues[grouperName] = value
			keyParts = append(keyParts, fmt.Sprintf("%v", value))
		}

		// Create unique key for this combination of grouper values
		key := strings.Join(keyParts, "|")

		// Accumulate duration
		aggregationMap[key] += record.duration
		grouperValuesMap[key] = grouperValues
	}

	// Convert map to slice of Aggregation structs
	aggregations := []Aggregation{}
	for key, duration := range aggregationMap {
		aggregations = append(aggregations, Aggregation{
			Groupers: grouperValuesMap[key],
			Duration: duration,
		})
	}

	return aggregations
}

// matchesFilters checks if a record matches all specified filters
func (a *App) matchesFilters(record Record, filters map[string]string) bool {
	for key, value := range filters {
		switch key {
		case "start_date":
			startDate := 0
			fmt.Sscanf(value, "%d", &startDate)
			if record.date_id < startDate {
				return false
			}
		case "end_date":
			endDate := 0
			fmt.Sscanf(value, "%d", &endDate)
			if record.date_id > endDate {
				return false
			}
		case "category":
			if record.category != value {
				return false
			}
		case "url":
			if !strings.Contains(record.url, value) {
				return false
			}
		case "exe_path":
			if record.exe_path != value {
				return false
			}
		case "name":
			if record.name != value {
				return false
			}
		case "is_weekend":
			isWeekend := value == "true"
			if record.date_info.IsWeekend != isWeekend {
				return false
			}
		}
	}
	return true
}

// extractGrouperValue extracts the value for a given grouper from a record
func (a *App) extractGrouperValue(record Record, grouper Grouper) interface{} {
	switch grouper {
	case GroupByDate:
		return record.date_id
	case GroupByWeek:
		return record.date_info.WeekOfYear
	case GroupByMonth:
		return (record.date_id / 100) % 100 // extract month from YYYYMMDD
	case GroupByYear:
		return record.date_id / 10000 // extract year from YYYYMMDD
	case GroupByDayOfWeek:
		return record.date_info.DayOfWeek
	case GroupByIsWeekend:
		return record.date_info.IsWeekend
	case GroupByCategory:
		return record.category
	case GroupByURL:
		return record.url
	case GroupByExePath:
		return record.exe_path
	case GroupByName:
		return record.name
	default:
		return nil
	}
}
