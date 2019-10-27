package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuloV/ics-golang"
	ical "github.com/arran4/golang-ical"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		calendarURL, err := getCalendarURL()
		if err != nil {
			http.Error(w, fmt.Sprintf("couldn't get calendar url: %v", err), 500)
			return
		}

		formattedCalendar, err := collectAndFormatCalendar(calendarURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("couldn't process calendar: %v", err), 500)
			return
		}

		fmt.Fprint(w, formattedCalendar)
		fmt.Printf("Requested at: %v\n", time.Now().Format("Jan 2, 2006 at 3:04pm (MST)"))
	})

	http.ListenAndServe(":8080", nil)
}

func getCalendarURL() (string, error) {
	if os.Getenv("CALENDAR_URL") != "" {
		fmt.Println("using calendar env var")
		return os.Getenv("CALENDAR_URL"), nil
	}

	data, err := ioutil.ReadFile(os.Getenv("CALENDAR_URL_PATH"))
	if err != nil {
		return "", fmt.Errorf("failed to load calendar from file: %v", err)
	}

	return strings.TrimSpace(string(data)), nil
}

func collectAndFormatCalendar(url string) (string, error) {
	parser := ics.New()
	inputChan := parser.GetInputChan()
	inputChan <- url

	parser.Wait()
	cal, err := parser.GetCalendars()
	if err != nil {
		return "", fmt.Errorf("failed to get cals: %s", err)
	}
	if len(cal) != 1 {
		return "", fmt.Errorf("unexpected number of cals (expected 1): %v", len(cal))
	}
	originalCalendar := cal[0]
	splitCalendar := ics.NewCalendar()
	splitCalendar.SetTimezone(originalCalendar.GetTimezone())

	for _, e := range originalCalendar.GetEvents() {
		eventSplitIndex := 0
		for {
			eventSplitIndex++
			start := e.GetStart()
			end := e.GetEnd()

			if start.Format("20060102") == end.Format("20060102") {
				splitCalendar.SetEvent(e)
				break
			}

			splitEvent := e
			s := e.GetStart()
			endTime := time.Date(s.Year(), s.Month(), s.Day(), 23, 59, 59, 0, s.Location())
			splitEvent.SetEnd(endTime)
			splitEvent.SetID(fmt.Sprintf("%s-%d", e.GetID(), eventSplitIndex))

			splitCalendar.SetEvent(splitEvent)

			newStart := endTime.Add(1 * time.Second)
			e.SetStart(newStart)
		}
	}

	outputCalendar := ical.NewCalendar()
	for _, e := range splitCalendar.GetEvents() {
		event := outputCalendar.AddEvent(e.GetID())
		event.SetCreatedTime(time.Now())
		event.SetDtStampTime(time.Now())
		event.SetStartAt(e.GetStart())
		event.SetEndAt(e.GetEnd())
		event.SetSummary(formatSummaryText(e.GetSummary()))
	}

	return outputCalendar.Serialize(), nil
}

func formatSummaryText(text string) string {
	if strings.Contains(text, "Back-up") {
		return "Backup"
	} else if strings.Contains(text, "On-call") {
		return "On Call"
	}
	return text
}
