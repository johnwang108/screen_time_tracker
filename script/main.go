package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"

	// vestigial imports for keyboard functionality
	_ "os/signal"
	_ "strconv"
	_ "syscall"

	_ "github.com/eiannone/keyboard"
)

type WindowReading struct {
	ExePath   string
	Timestamp time.Time
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

	return WindowReading{
		ExePath:   exeName,
		Timestamp: time.Now(),
	}, nil
}

// storeReading persists a window reading
func storeReading(reading WindowReading) {

	print("Focused: ", reading.ExePath, " at ", reading.Timestamp.Format(time.RFC3339), "\n")
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
		writer.Write([]string{"name", "timestamp"})
	}

	writer.Write([]string{reading.ExePath, reading.Timestamp.Format(time.RFC3339)})
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

		duration := nextTime.Sub(currentTime)
		result[exePath] += duration.Minutes()
	}

	return result, nil
}

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	// Simple 16x16 red square ICO as placeholder
	icon := []byte{
		0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x10, 0x10, 0x00, 0x00, 0x01, 0x00,
		0x18, 0x00, 0x68, 0x03, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x28, 0x00,
		0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x01, 0x00,
		0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00,
	}
	// Fill with blue pixels (BGR format)
	for i := 0; i < 16*16; i++ {
		icon = append(icon, 0xFF, 0x00, 0x00) // Blue in BGR
	}
	// Add mask (all opaque)
	for i := 0; i < 16*16/8; i++ {
		icon = append(icon, 0x00)
	}

	systray.SetIcon(icon)
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

// TabInfo represents the data received from the browser extension
type TabInfo struct {
	TabID int64  `json:"tabId"`
	Title string `json:"title"`
	URL   string `json:"url"`
	TS    int64  `json:"ts"`
}

// NativeMessage is the structure for sending messages back to the extension
type NativeMessage struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func main() {
	// Set up logging to a file (can't use stdout - that's for native messaging)
	logFile, err := os.OpenFile("/tmp/tabhost.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// If we can't open log file, fail silently
		return
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Println("Tab tracker native host started")

	// Read messages from stdin (sent by the browser extension)
	for {
		msg, err := readMessage(os.Stdin)
		if err != nil {
			if err == io.EOF {
				log.Println("Extension disconnected (EOF)")
				break
			}
			log.Printf("Error reading message: %v\n", err)
			continue
		}

		// Parse the JSON message
		var tabInfo TabInfo
		if err := json.Unmarshal(msg, &tabInfo); err != nil {
			log.Printf("Error parsing JSON: %v\n", err)
			sendError(fmt.Sprintf("Invalid JSON: %v", err))
			continue
		}

		// Process the tab info (this is where you'd add your custom logic)
		handleTabInfo(tabInfo)

	}
}

// 		// Optionally send a response back to the extension
// 		sendResponse(NativeMessage{Status: "ok"})
// 	}

// 	log.Println("Tab tracker native host stopped")
// }

// // readMessage reads a single message from the extension using native messaging protocol
// // Messages are length-prefixed: 4 bytes (uint32 little-endian) followed by the message
// func readMessage(r io.Reader) ([]byte, error) {
// 	// Read the 4-byte message length prefix
// 	var length uint32
// 	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
// 		return nil, err
// 	}

// 	// Sanity check: prevent reading massive messages
// 	if length > 1024*1024 { // 1MB limit
// 		return nil, fmt.Errorf("message too large: %d bytes", length)
// 	}

// 	// Read the actual message
// 	msg := make([]byte, length)
// 	if _, err := io.ReadFull(r, msg); err != nil {
// 		return nil, err
// 	}

// 	return msg, nil
// }

// // writeMessage sends a message to the extension using native messaging protocol
// func writeMessage(w io.Writer, msg []byte) error {
// 	// Write the 4-byte length prefix
// 	length := uint32(len(msg))
// 	if err := binary.Write(w, binary.LittleEndian, length); err != nil {
// 		return err
// 	}

// 	// Write the actual message
// 	if _, err := w.Write(msg); err != nil {
// 		return err
// 	}

// 	return nil
// }

// // handleTabInfo processes the received tab information
// // Customize this function with your own logic
// func handleTabInfo(info TabInfo) {
// 	timestamp := time.Unix(info.TS/1000, (info.TS%1000)*1000000)

// 	log.Printf("Tab Info Received:")
// 	log.Printf("  Tab ID: %d", info.TabID)
// 	log.Printf("  Title:  %s", info.Title)
// 	log.Printf("  URL:    %s", info.URL)
// 	log.Printf("  Time:   %s", timestamp.Format("2006-01-02 15:04:05"))
// 	log.Println()

// 	// Example: Save to a file
// 	saveToFile(info)

// 	// Example: Send to a database, API, etc.
// 	// db.Save(info)
// 	// api.Post("/tabs", info)
// }

// // saveToFile appends tab info to a CSV file
// func saveToFile(info TabInfo) {
// 	f, err := os.OpenFile("/tmp/tab_history.csv", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
// 	if err != nil {
// 		log.Printf("Error opening file: %v\n", err)
// 		return
// 	}
// 	defer f.Close()

// 	// Check if file is empty (needs header)
// 	stat, _ := f.Stat()
// 	if stat.Size() == 0 {
// 		f.WriteString("timestamp,tab_id,title,url\n")
// 	}

// 	timestamp := time.Unix(info.TS/1000, 0).Format("2006-01-02 15:04:05")
// 	// Escape commas and quotes in CSV
// 	title := fmt.Sprintf("\"%s\"", info.Title)
// 	url := fmt.Sprintf("\"%s\"", info.URL)

// 	line := fmt.Sprintf("%s,%d,%s,%s\n", timestamp, info.TabID, title, url)
// 	f.WriteString(line)
// }

// // sendResponse sends a message back to the extension
// func sendResponse(msg NativeMessage) {
// 	data, err := json.Marshal(msg)
// 	if err != nil {
// 		log.Printf("Error marshaling response: %v\n", err)
// 		return
// 	}

// 	if err := writeMessage(os.Stdout, data); err != nil {
// 		log.Printf("Error writing response: %v\n", err)
// 	}
// }

// // sendError sends an error message back to the extension
// func sendError(errMsg string) {
// 	sendResponse(NativeMessage{
// 		Status: "error",
// 		Error:  errMsg,
// 	})
// }
