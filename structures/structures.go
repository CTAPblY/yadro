package structures

import "time"

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
