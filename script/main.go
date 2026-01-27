package main

import (
	_ "embed"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/getlantern/systray"
	hook "github.com/robotn/gohook"
	"golang.org/x/sys/windows"

	// vestigial imports for keyboard functionality
	_ "os/signal"
	_ "strconv"
	_ "syscall"

	_ "github.com/eiannone/keyboard"
)

//go:embed timer.ico
var iconBytes []byte

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

	hadActivitySinceLastCheck bool
	activityMu                sync.Mutex
)

type WindowReading struct {
	ExePath     string
	TabName     string
	TabUrl      string
	Timestamp   time.Time
	HadActivity bool
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
	// Check for activity since last call
	hadActivity := checkAndResetActivity()

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
		ExePath:     exeName,
		TabName:     tabName,
		TabUrl:      tabUrl,
		Timestamp:   time.Now(),
		HadActivity: hadActivity,
	}, nil
}

// startActivityMonitor starts a global event hook to detect mouse and keyboard activity
func startActivityMonitor() {
	evChan := hook.Start()
	defer hook.End()

	for range evChan {
		activityMu.Lock()
		hadActivitySinceLastCheck = true
		activityMu.Unlock()
	}
}

// checkAndResetActivity checks if there was activity since last check and resets the flag
func checkAndResetActivity() bool {
	activityMu.Lock()
	defer activityMu.Unlock()

	result := hadActivitySinceLastCheck
	hadActivitySinceLastCheck = false
	return result
}

// storeReading persists a window reading
func storeReading(reading WindowReading) {
	// create data folder in AppData/Local
	data_dir := filepath.Join(os.Getenv("LOCALAPPDATA"), "tracker_data")
	err := os.MkdirAll(data_dir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// create or append to file named by date
	today := time.Now().Format("20060102")
	data_file_path := filepath.Join(data_dir, today+".csv")

	// check if file exists to determine if we need headers
	isNew := false
	if _, err := os.Stat(data_file_path); os.IsNotExist(err) {
		isNew = true
	}

	f, err := os.OpenFile(data_file_path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// write header if new file
	if isNew {
		writer.Write([]string{"name", "timestamp", "tabName", "tabUrl", "hadActivity"})
	}

	writer.Write([]string{reading.ExePath, reading.Timestamp.Format(time.RFC3339), reading.TabName, reading.TabUrl, fmt.Sprintf("%t", reading.HadActivity)})
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(iconBytes)
	systray.SetTitle("Tracker")
	systray.SetTooltip("Window Tracker")

	mQuit := systray.AddMenuItem("Exit", "Exit the tracker")

	// Handle quit menu click
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

	// Start activity monitor
	go startActivityMonitor()

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
