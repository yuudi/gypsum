package gypsum


type trigger struct {
	GroupID     int64       `json:"group_id"`
	UserID      int64       `json:"user_id"`
	TriggerType string `json:"trigger_type"`
}