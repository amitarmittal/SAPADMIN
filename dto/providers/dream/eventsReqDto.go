package dream

type EventsReqDto struct {
	SportId         string   `json:"sportId"`
	CompetitionsIds []string `json:"competitionsIds"`
}
