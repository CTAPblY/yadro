package loader

import (
	"os"
	"strconv"
	"strings"

	"github.com/CTAPblY/yadro/structures"
)

func LoadEvents(filename string) ([]structures.Event, error) {
	var (
		ExtraParams string
		events      []structures.Event
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
		events = append(events, structures.Event{
			Time:         time,
			EventID:      EventID,
			CompetitorID: CompetitorID,
			ExtraParams:  ExtraParams,
		})
	}

	return events, nil
}
