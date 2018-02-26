package api

type tbaEvent struct {
	Key       string   `json:"key"`
	Name      string   `json:"name"`
	ShortName string   `json:"short_name"`
	EventType int      `json:"event_type"`
	Lat       *float64 `json:"lat"`
	Lng       *float64 `json:"lng"`
	Date      string   `json:"start_date"`
	EndDate   string   `json:"end_date"`
	TimeZone  string   `json:"timezone"`
}

type tbaMatch struct {
	Key             string `json:"key"`
	ScheduledTime   int64  `json:"time"`
	PredictedTime   int64  `json:"predicted_time"`
	ActualTime      int64  `json:"actual_time"`
	WinningAlliance string `json:"winning_alliance"`
	Alliances       struct {
		Blue struct {
			Score int      `json:"score"`
			Teams []string `json:"team_keys"`
		} `json:"blue"`
		Red struct {
			Score int      `json:"score"`
			Teams []string `json:"team_keys"`
		} `json:"red"`
	} `json:"alliances"`
}

type media struct {
	Type       string `json:"type"`
	ForeignKey string `json:"foreign_key"`
	Details    struct {
		ThumbnailURL string `json:"thumbnail_url"`
	} `json:"details"`
}
