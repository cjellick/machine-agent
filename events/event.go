package events

type Event struct {
	Name  string `json:"name"`
	Sleep int    `json:"sleep"`
	Data  string `json:"data"`
}
