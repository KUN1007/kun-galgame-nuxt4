package dto

type CalendarMeta struct {
	PrevMonth string `json:"prev_month"`
	NextMonth string `json:"next_month"`
	HasPrev   bool   `json:"has_prev"`
	HasNext   bool   `json:"has_next"`
	MinMonth  string `json:"min_month"`
	MaxMonth  string `json:"max_month"`
	Count     int    `json:"count"`
}

type CalendarMonthPage struct {
	Month string        `json:"month"`
	Today string        `json:"today"`
	Items []GalgameCard `json:"items"`
	Meta  CalendarMeta  `json:"meta"`
}

type CalendarPendingPage struct {
	Year  string        `json:"year"`
	Items []GalgameCard `json:"items"`
	Count int           `json:"count"`
}

type CalendarTBAPage struct {
	Items []GalgameCard `json:"items"`
	Count int           `json:"count"`
}

type CalendarUpcomingMonth struct {
	Month string        `json:"month"`
	Items []GalgameCard `json:"items"`
}

type CalendarUpcomingPage struct {
	Today  string                  `json:"today"`
	Months []CalendarUpcomingMonth `json:"months"`
	Count  int                     `json:"count"`
}

type CalendarTodayFlag struct {
	Today      string `json:"today"`
	HasRelease bool   `json:"has_release"`
	// Seconds until the flag can change. The sidebar keeps the answer in local
	// storage for that long instead of pulling a month of cards on every load.
	ExpiresIn int `json:"expires_in"`
}
