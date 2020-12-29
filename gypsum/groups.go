package gypsum

type ItemType string

const (
	RuleItem      ItemType = "rule"
	TriggerItem   ItemType = "trigger"
	SchedulerItem ItemType = "scheduler"
	ResourceItem  ItemType = "resource"
	GroupItem     ItemType = "group"
)

type Group struct {
	DisplayName string `json:"display_name"`
	Items       []struct {
		Type ItemType `json:"type"`
		ID   uint64   `json:"id"`
	} `json:"items"`
}
