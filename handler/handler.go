package handler

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/CTAPblY/yadro/structures"
)

func HandleEvents(config *structures.Config, events []structures.Event) ([]structures.Result, error) {
	competitors := make(map[int]*structures.Competitor)
	var results []structures.Result
	for _, event := range events {
		competitor, exists := competitors[event.CompetitorID]
		if !exists {
			competitor = &structures.Competitor{
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
		result := structures.Result{
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
