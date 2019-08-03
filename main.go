package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(os.Getenv("CALENDAR_URL"))
		if err != nil {
			http.Error(w, fmt.Sprintf("could not get calendar: %v", err), 500)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprint("failed to read calendar resp body", err), 500)
			return
		}

		fmt.Fprint(w, formatSummaryText(string(body)))
		fmt.Printf("Requested at: %v\n", time.Now().Format("Jan 2, 2006 at 3:04pm (MST)"))
	})

	http.ListenAndServe(":8080", nil)
}

func formatSummaryText(calendarData string) string {
	lines := strings.Split(calendarData, "\n")

	for i, v := range lines {
		if strings.HasPrefix(v, "SUMMARY:") {
			if strings.Contains(v, "Back-up") {
				lines[i] = "SUMMARY: Backup"
			} else {
				lines[i] = "SUMMARY: On Call"
			}
		}
	}

	return strings.Join(lines, "\n")
}
