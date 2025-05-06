package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Laps        int    `json:"laps"`
	LapLen      int    `json:"lapLen"`
	PenaltyLen  int    `json:"penaltyLen"`
	FiringLines int    `json:"firingLines"`
	Statrt      string `json:"start"`
	StartDelta  string `json:"startDelta"`
}

type Event struct {
	Time         string
	EventID      int
	CompetitorID int
	ExtraParams  string
}

type Result struct {
	ID           int
	Status       string
	LapTimes     []string
	LapSpeed     []string
	PenaltyLen   string
	PenaltySpeed string
	Hits         string
}

type Competitor struct {
	ID             int
	Registered     bool
	ScheduledStart time.Time
	ActualStart    time.Time
	Finished       bool
	Started        bool
	NotFinished    bool
	Comment        string
	LapTimes       []time.Duration
	PenaltyTime    time.Duration
	Hits           int
	Shots          int
	CurrentLap     int
	OnFiringRange  bool
	OnPenalty      bool
	CurrentLapTime time.Time
	LastEventTime  time.Time
}

func main() {
	config, err := LoadConfig("config.json")

	if err != nil {
		fmt.Println(err)
	}
	events, err := LoadEvents("events")
	if err != nil {
		fmt.Println(err)
	}
	report, err := HandleEvents(config, events)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(report)
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func LoadEvents(filename string) ([]Event, error) {
	var (
		ExtraParams string
		events      []Event
	)
	file, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		values := strings.Split(line, " ")
		time := strings.Trim(values[0], "[]")

		EventID, err := strconv.Atoi(values[1])
		if err != nil {
			return nil, err
		}

		CompetitorID, err := strconv.Atoi(values[2])
		if err != nil {
			return nil, err
		}

		if len(values) > 3 {
			ExtraParams = values[3]
		}
		events = append(events, Event{
			Time:         time,
			EventID:      EventID,
			CompetitorID: CompetitorID,
			ExtraParams:  ExtraParams,
		})
	}

	return events, nil
}

func HandleEvents(config *Config, events []Event) ([]Result, error) {
	competitors := make(map[int]*Competitor)
	// var outputEvents []string
	var results []Result
	for _, event := range events {
		competitor, exists := competitors[event.CompetitorID]
		if !exists {
			competitor = &Competitor{
				ID:         event.CompetitorID,
				LapTimes:   make([]time.Duration, config.Laps),
				CurrentLap: 0,
			}
			competitors[event.CompetitorID] = competitor
		}

		switch event.EventID {
		case 1:
			competitor.Registered = true
			time, err := ParseTime(event.Time)
			if err != nil {
				return nil, err
			}
			competitor.LastEventTime = time
			fmt.Println("[", event.Time, "] The competitor(", event.CompetitorID, ") registered")
		case 2:
			startTime, err := ParseTime(event.ExtraParams)
			if err != nil {
				return nil, err
			}
			time, err := ParseTime(event.Time)
			if err != nil {
				return nil, err
			}
			competitor.LastEventTime = time
			competitor.ScheduledStart = startTime
			fmt.Println("[", event.Time, "] The start time for compettitor(", event.CompetitorID, ") was set by a draw to ", event.ExtraParams)
		case 3:
			time, err := ParseTime(event.Time)
			if err != nil {
				return nil, err
			}
			competitor.LastEventTime = time
			fmt.Println("[", event.Time, "] The competitor(", event.CompetitorID, ") is on the start line")
		case 4:
			StartInTime, err := CompetitorStartInTime(config.Statrt, config.StartDelta, event.Time)
			if err != nil {
				return nil, err
			}
			if StartInTime {
				competitor.Started = true
				startTime, err := ParseTime(event.Time)
				if err != nil {
					return nil, err
				}

				competitor.CurrentLapTime = startTime
				competitor.LastEventTime = startTime
				competitor.CurrentLap = 0

				fmt.Println("[", event.Time, "] The competitior(", event.CompetitorID, ") has started")
			} else {
				competitor.Started = false
			}
		case 5:
			competitor.OnFiringRange = true
			time, err := ParseTime(event.Time)
			if err != nil {
				return nil, err
			}
			competitor.LastEventTime = time
			fmt.Println("[", event.Time, "] The competitor(", event.CompetitorID, ") on the firing range(", event.ExtraParams, ")")
		case 6:
			competitor.Hits++
			time, err := ParseTime(event.Time)
			if err != nil {
				return nil, err
			}
			competitor.LastEventTime = time
			fmt.Println("[", event.Time, "] The target(", event.ExtraParams, ") has been hit by competitor(", event.CompetitorID, ")")
		case 7:
			competitor.OnFiringRange = false
			time, err := ParseTime(event.Time)
			if err != nil {
				return nil, err
			}
			competitor.LastEventTime = time
			fmt.Println("[", event.Time, "] The competitor(", event.CompetitorID, ") has left the firing range")
		case 8:
			competitor.OnPenalty = true
			time, err := ParseTime(event.Time)
			if err != nil {
				return nil, err
			}
			competitor.LastEventTime = time
			fmt.Println("[", event.Time, "] The competitor(", event.CompetitorID, ") entered the penalty laps")
		case 9:
			time, err := ParseTime(event.Time)
			if err != nil {
				return nil, err
			}
			if competitor.OnPenalty {
				competitor.PenaltyTime += time.Sub(competitor.LastEventTime)
				competitor.OnPenalty = false
			}
			competitor.LastEventTime = time
			fmt.Println("[", event.Time, "] The competitor(", event.CompetitorID, ") left the penalty laps")
		case 10:
			time, err := ParseTime(event.Time)
			if err != nil {
				return nil, err
			}
			if competitor.CurrentLap < config.Laps {
				lapDuration := time.Sub(competitor.CurrentLapTime)
				competitor.LapTimes[competitor.CurrentLap] = lapDuration

				competitor.CurrentLapTime = time
				competitor.CurrentLap++
			}
			competitor.LastEventTime = time
			fmt.Println("[", event.Time, "] THe competitor(", event.CompetitorID, ") ended the main lap")
		case 11:
			time, err := ParseTime(event.Time)
			if err != nil {
				return nil, err
			}
			competitor.NotFinished = true
			competitor.Comment = event.ExtraParams
			competitor.LastEventTime = time
			fmt.Println("[", event.Time, "] The competitor(", event.CompetitorID, ") can't continue: ", event.ExtraParams)
		}
	}
	for _, competitor := range competitors {
		result := Result{
			ID:   competitor.ID,
			Hits: fmt.Sprintf("%d/%d", competitor.Hits, config.FiringLines*5),
		}

		switch {
		case !competitor.Started:
			result.Status = "NotStarted"
		case competitor.NotFinished:
			result.Status = "NotFinished"
		case competitor.CurrentLap == config.Laps:
			result.Status = "Finished"

			result.LapTimes = make([]string, config.Laps)
			result.LapSpeed = make([]string, config.Laps)
			for i := 0; i < config.Laps; i++ {
				if i < len(competitor.LapTimes) {
					lapTime := competitor.LapTimes[i]
					result.LapTimes[i] = formatDuration(lapTime)
					if lapTime > 0 {
						result.LapSpeed[i] = fmt.Sprintf("%.3f", float64(config.LapLen)/lapTime.Seconds())
					} else {
						result.LapSpeed[i] = "0.000"
					}
				} else {
					result.LapTimes[i] = "00:00:00.000"
					result.LapSpeed[i] = "0.000"
				}
			}

			result.PenaltyLen = formatDuration(competitor.PenaltyTime)
			if competitor.PenaltyTime > 0 {
				result.PenaltySpeed = fmt.Sprintf("%.3f", float64(config.PenaltyLen)/competitor.PenaltyTime.Seconds())
			} else {
				result.PenaltySpeed = "0.000"
			}
		default:
			continue
		}

		results = append(results, result)
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Status == results[j].Status {
			return results[i].ID < results[j].ID
		}
		return results[i].Status == "Finished" ||
			(results[i].Status == "NotFinished" && results[j].Status == "NotStarted")
	})

	return results, nil
}

func ParseTime(timeStr string) (time.Time, error) {
	parsedTime, err := time.Parse("15:04:05.000", timeStr)
	return parsedTime, err
}

func CompetitorStartInTime(startTimeStr string, startDeltaStr string, actualStartStr string) (bool, error) {
	startTime, err := ParseTime(startTimeStr)
	if err != nil {
		return false, err
	}

	delta, err := parseDelta(startDeltaStr)
	if err != nil {
		return false, err
	}

	actualStart, err := ParseTime(actualStartStr)
	if err != nil {
		return false, err
	}

	maxStartTime := startTime.Add(delta)

	return !actualStart.After(maxStartTime), nil
}

func parseDelta(deltaStr string) (time.Duration, error) {
	parts := strings.Split(deltaStr, ":")

	hours, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, err
	}

	secondsParts := strings.Split(parts[2], ".")
	secs, err := strconv.Atoi(secondsParts[0])
	if err != nil {
		return 0, err
	}

	nanos := 0
	if len(secondsParts) > 1 {
		ms, err := strconv.Atoi(secondsParts[1])
		if err != nil {
			return 0, err
		}
		nanos = ms * 1e6
	}

	return time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(secs)*time.Second +
		time.Duration(nanos)*time.Nanosecond, nil
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	millis := int(d.Milliseconds()) % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, millis)
}
