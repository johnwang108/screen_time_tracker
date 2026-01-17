package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"

	// vestigial imports for keyboard functionality
	_ "os/signal"
	_ "strconv"
	_ "syscall"

	_ "github.com/eiannone/keyboard"
)

const (
	httpPort          = "8384"
	chromeExtensionID = "YOUR_CHROME_EXTENSION_ID"
	// firefoxExtensionID   = "@jwtabtracker_local"
	firefoxExtensionID = "583a7b7a-defb-4431-8771-cee0ca64931a"

	allowedOriginChrome  = "chrome-extension://" + chromeExtensionID
	allowedOriginFirefox = "moz-extension://" + firefoxExtensionID
)

var (
	currentTab   TabInfo
	currentTabMu sync.Mutex
)

type WindowReading struct {
	ExePath   string
	TabName   string
	TabUrl    string
	Timestamp time.Time
}

// TabInfo represents the data received from the browser extension
type TabInfo struct {
	TabID int64  `json:"tabId"`
	Title string `json:"title"`
	URL   string `json:"url"`
	TS    int64  `json:"ts"`
}

// getFocusedWindowInfo retrieves the exe path of the currently focused window
func getFocusedWindowInfo() (WindowReading, error) {
	hwnd := windows.GetForegroundWindow()

	// Get process ID from window handle
	var pid uint32
	_, _ = windows.GetWindowThreadProcessId(hwnd, &pid)

	// Get executable path from PID
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return WindowReading{}, err
	}
	defer windows.CloseHandle(handle)

	var buf [windows.MAX_PATH]uint16
	length := uint32(len(buf))
	err = windows.QueryFullProcessImageName(handle, 0, &buf[0], &length)
	if err != nil {
		return WindowReading{}, err
	}
	exePath := windows.UTF16ToString(buf[:length])
	parts := strings.Split(exePath, "\\")
	exeName := parts[len(parts)-1]

	browserNames := []string{"chrome.exe", "firefox.exe"}
	tabName := ""
	tabUrl := ""

	for _, v := range browserNames {
		if exeName == v {
			currentTabMu.Lock()
			tabName = currentTab.Title
			tabUrl = currentTab.URL
			currentTabMu.Unlock()
			break
		}
	}

	return WindowReading{
		ExePath:   exeName,
		TabName:   tabName,
		TabUrl:    tabUrl,
		Timestamp: time.Now(),
	}, nil
}

// storeReading persists a window reading
func storeReading(reading WindowReading) {
	// create data folder
	err := os.MkdirAll("data/", 0755)
	if err != nil {
		log.Fatal(err)
	}

	// create or append to file named by date
	today := time.Now().Format("20060102")
	filepath := "data/" + today + ".csv"

	// check if file exists to determine if we need headers
	isNew := false
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		isNew = true
	}

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// write header if new file
	if isNew {
		writer.Write([]string{"name", "timestamp", "tabName", "tabUrl"})
	}

	writer.Write([]string{reading.ExePath, reading.Timestamp.Format(time.RFC3339), reading.TabName, reading.TabUrl})
}

// calculateTimeSpent reads the CSV file for the given date and returns
// the time spent (in minutes) with each application focused.
func calculateTimeSpent(date int) (map[string]float64, error) {
	filepath := fmt.Sprintf("data/%d.csv", date)

	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// skip header row, need at least 2 data rows to calculate duration
	if len(records) < 3 {
		return map[string]float64{}, nil
	}

	result := make(map[string]float64)

	// iterate through consecutive pairs of readings
	for i := 0; i < len(records)-1; i++ {
		exePath := records[i][0]
		if exePath == "Off" {
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

		for duration := nextTime.Sub(currentTime); duration.Seconds() <= 15; { // if longer than 15 seconds, likely computer was off or asleep
			result[exePath] += duration.Minutes()
		}
	}

	return result, nil
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	// Simple 16x16 red square ICO as placeholder
	systray.SetTitle("Tracker")
	systray.SetTooltip("Window Tracker")

	mQuit := systray.AddMenuItem("Exit", "Exit the tracker")

	// Handle quit menu click
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

	// Start tracking loop
	go trackingLoop()

	// Start HTTP server for browser extension
	go startHTTPServer()
}

func onExit() {
	storeReading(WindowReading{
		ExePath:   "Off",
		Timestamp: time.Now(),
	})
}

func trackingLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Take an initial reading immediately
	if reading, err := getFocusedWindowInfo(); err == nil {
		storeReading(reading)
	}

	for {
		<-ticker.C
		reading, err := getFocusedWindowInfo()
		if err != nil {
			continue
		}
		storeReading(reading)
	}
}

func startHTTPServer() {
	http.HandleFunc("/tab", tabHandler)
	addr := "127.0.0.1:" + httpPort
	log.Printf("Starting HTTP server on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("HTTP server error: %v", err)
	}
}

func tabHandler(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// Validate origin
	if origin != allowedOriginChrome && origin != allowedOriginFirefox {
		http.Error(w, "Forbidden", http.StatusForbidden)
		log.Printf("Wrong origin: %v", origin)
		return
	}

	print("Received tab info from origin: ", origin, "\n")

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Handle preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tabInfo TabInfo
	if err := json.NewDecoder(r.Body).Decode(&tabInfo); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	currentTabMu.Lock()
	currentTab = tabInfo
	currentTabMu.Unlock()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
