package main

import (
	"context"
	"fmt"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called at application startup
func (a *App) startup(ctx context.Context) {
	// Perform your setup here
	a.ctx = ctx
}

// domReady is called after front-end resources have been loaded
func (a App) domReady(ctx context.Context) {
	// Add your action here
}

// beforeClose is called when the application is about to quit,
// either by clicking the window close button or calling runtime.Quit.
// Returning true will cause the application to continue, false will continue shutdown as normal.
func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	return false
}

// shutdown is called at application termination
func (a *App) shutdown(ctx context.Context) {
	// Perform your teardown here
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// type WindowReading struct {
// 	ExePath   string
// 	Timestamp time.Time
// }

// // getFocusedWindowInfo retrieves the exe path of the currently focused window
// func getFocusedWindowInfo() (WindowReading, error) {
// 	hwnd := windows.GetForegroundWindow()

// 	// Get process ID from window handle
// 	var pid uint32
// 	_, _ = windows.GetWindowThreadProcessId(hwnd, &pid)

// 	// Get executable path from PID
// 	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
// 	if err != nil {
// 		return WindowReading{}, err
// 	}
// 	defer windows.CloseHandle(handle)

// 	var buf [windows.MAX_PATH]uint16
// 	length := uint32(len(buf))
// 	err = windows.QueryFullProcessImageName(handle, 0, &buf[0], &length)
// 	if err != nil {
// 		return WindowReading{}, err
// 	}
// 	exePath := windows.UTF16ToString(buf[:length])
// 	parts := strings.Split(exePath, "\\")
// 	exeName := parts[len(parts)-1]

// 	return WindowReading{
// 		ExePath:   exeName,
// 		Timestamp: time.Now(),
// 	}, nil
// }

// // storeReading persists a window reading
// func storeReading(reading WindowReading) {

// 	print("Focused: ", reading.ExePath, " at ", reading.Timestamp.Format(time.RFC3339), "\n")
// 	// create data folder
// 	err := os.MkdirAll("data/", 0755)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// create or append to file named by date
// 	today := time.Now().Format("20060102")
// 	filepath := "data/" + today + ".csv"

// 	// check if file exists to determine if we need headers
// 	isNew := false
// 	if _, err := os.Stat(filepath); os.IsNotExist(err) {
// 		isNew = true
// 	}

// 	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer f.Close()

// 	writer := csv.NewWriter(f)
// 	defer writer.Flush()

// 	// write header if new file
// 	if isNew {
// 		writer.Write([]string{"name", "timestamp"})
// 	}

// 	writer.Write([]string{reading.ExePath, reading.Timestamp.Format(time.RFC3339)})
// }

// // calculateTimeSpent reads the CSV file for the given date and returns
// // the time spent (in minutes) with each application focused.
// func calculateTimeSpent(date int) (map[string]float64, error) {
// 	filepath := fmt.Sprintf("data/%d.csv", date)

// 	f, err := os.Open(filepath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer f.Close()

// 	reader := csv.NewReader(f)
// 	records, err := reader.ReadAll()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// skip header row, need at least 2 data rows to calculate duration
// 	if len(records) < 3 {
// 		return map[string]float64{}, nil
// 	}

// 	result := make(map[string]float64)

// 	// iterate through consecutive pairs of readings
// 	for i := 0; i < len(records)-1; i++ {
// 		exePath := records[i][0]
// 		if exePath == "Off" {
// 			continue
// 		}
// 		currentTime, err := time.Parse(time.RFC3339, records[i][1])
// 		if err != nil {
// 			continue
// 		}
// 		nextTime, err := time.Parse(time.RFC3339, records[i+1][1])
// 		if err != nil {
// 			continue
// 		}

// 		duration := nextTime.Sub(currentTime)
// 		result[exePath] += duration.Minutes()
// 	}

// 	return result, nil
// }

// func main() {
// 	systray.Run(onReady, onExit)
// }

// func onReady() {
// 	// Simple 16x16 red square ICO as placeholder
// 	icon := []byte{
// 		0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x10, 0x10, 0x00, 0x00, 0x01, 0x00,
// 		0x18, 0x00, 0x68, 0x03, 0x00, 0x00, 0x16, 0x00, 0x00, 0x00, 0x28, 0x00,
// 		0x00, 0x00, 0x10, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x01, 0x00,
// 		0x18, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00,
// 		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
// 		0x00, 0x00,
// 	}
// 	// Fill with blue pixels (BGR format)
// 	for i := 0; i < 16*16; i++ {
// 		icon = append(icon, 0xFF, 0x00, 0x00) // Blue in BGR
// 	}
// 	// Add mask (all opaque)
// 	for i := 0; i < 16*16/8; i++ {
// 		icon = append(icon, 0x00)
// 	}

// 	systray.SetIcon(icon)
// 	systray.SetTitle("Tracker")
// 	systray.SetTooltip("Window Tracker")

// 	mQuit := systray.AddMenuItem("Exit", "Exit the tracker")

// 	// Handle quit menu click
// 	go func() {
// 		<-mQuit.ClickedCh
// 		systray.Quit()
// 	}()

// 	// Start tracking loop
// 	go trackingLoop()
// }

// func onExit() {
// 	storeReading(WindowReading{
// 		ExePath:   "Off",
// 		Timestamp: time.Now(),
// 	})
// }

// func trackingLoop() {
// 	ticker := time.NewTicker(5 * time.Second)
// 	defer ticker.Stop()

// 	// Take an initial reading immediately
// 	if reading, err := getFocusedWindowInfo(); err == nil {
// 		storeReading(reading)
// 	}

// 	for {
// 		<-ticker.C
// 		reading, err := getFocusedWindowInfo()
// 		if err != nil {
// 			continue
// 		}
// 		storeReading(reading)
// 	}
// }
