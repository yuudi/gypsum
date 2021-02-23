package gypsum

import (
	"encoding/gob"
	"errors"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
)

type ItemType string

const (
	RuleItem      ItemType = "rule"
	TriggerItem   ItemType = "trigger"
	SchedulerItem ItemType = "scheduler"
	ResourceItem  ItemType = "resource"
	GroupItem     ItemType = "group"
)

type UserRecord interface {
	ToBytes() ([]byte, error)
	GetParentID() uint64
	GetDisplayName() string
	NewParent(selfID, parentID uint64) error
	SaveToDB(selfID uint64) error
}

type RuleDescriptor interface {
	ToRule() zero.Rule
}

func init() {
	gob.Register(Group{})
	gob.Register(Job{})
	gob.Register(Resource{})
	gob.Register(Rule{})
	gob.Register(Trigger{})
}

func RuleAlwaysTrue(_ *zero.Event, _ zero.State) bool {
	return true
}

func RuleAlwaysFalse(_ *zero.Event, _ zero.State) bool {
	return false
}

func RestoreFromUserRecord(itemType ItemType, itemBytes []byte, newParentID uint64) (uint64, error) {
	switch itemType {
	case RuleItem:
		rule, err := RuleFromBytes(itemBytes)
		if err != nil {
			return 0, err
		}
		rule.ParentGroup = newParentID
		cursor := itemCursor.Require()
		rules[cursor] = rule
		if err := rule.SaveToDB(cursor); err != nil {
			return 0, err
		}
		return cursor, nil
	case TriggerItem:
		trigger, err := TriggerFromByte(itemBytes)
		if err != nil {
			return 0, err
		}
		trigger.ParentGroup = newParentID
		cursor := itemCursor.Require()
		triggers[cursor] = trigger
		if err := trigger.SaveToDB(cursor); err != nil {
			return 0, err
		}
		return cursor, nil
	case SchedulerItem:
		job, err := JobFromBytes(itemBytes)
		if err != nil {
			return 0, err
		}
		job.ParentGroup = newParentID
		cursor := itemCursor.Require()
		jobs[cursor] = job
		if err := job.SaveToDB(cursor); err != nil {
			return 0, err
		}
		return cursor, nil
	case ResourceItem:
		resource, err := ResourceFromBytes(itemBytes)
		if err != nil {
			return 0, err
		}
		resource.ParentGroup = newParentID
		cursor := itemCursor.Require()
		resources[cursor] = resource
		if err := resource.SaveToDB(cursor); err != nil {
			return 0, err
		}
		return cursor, nil
	case GroupItem:
		err := errors.New("group in group are not supported yet")
		log.Warn(err)
		return 0, err
	default:
		err := errors.New("unexpected type of user_record")
		log.Warnf("unknown type: %s", itemType)
		return 0, err
	}
}
