package gypsum

import (
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"

	"github.com/yuudi/gypsum/gypsum/helper"
)

type LimiterType string

const (
	// UnitTimeLimiterType limit usage per unit time
	UnitTimeLimiterType LimiterType = "unit_time"
	// DurationLimiterType limit usage at interval
	DurationLimiterType LimiterType = "duration"
)

type LimiterDescriptor struct {
	LimiterType    LimiterType `json:"limiter_type"`
	Unit           string      `json:"unit"`
	DurationSecond int32       `json:"duration_second"`
	MaxUsage       uint32      `json:"max_usage"`
	databaseKey    []byte
}

type Limiter interface {
	SaveToDB(uint64) error
	Require() bool
}

type durationLimiter struct {
	duration   time.Duration
	loopRecord []int64
	maxUsage   uint32
	cursor     uint32
	lock       sync.Mutex
}

func NewDurationLimiter(durationSecond int32, maxUsage uint32) *durationLimiter {
	return &durationLimiter{
		duration:   time.Duration(durationSecond) * time.Second,
		loopRecord: make([]int64, maxUsage),
		maxUsage:   maxUsage,
	}
}

func (d *durationLimiter) Require() bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	now := time.Now().UnixNano()
	notUntil := d.loopRecord[d.cursor]
	if now > notUntil {
		d.cursor += 1
		if d.cursor >= d.maxUsage {
			d.cursor = 0
		}
		d.loopRecord[d.cursor] = now + d.duration.Nanoseconds()
		return true
	} else {
		return false
	}
}

func (d *durationLimiter) SaveToDB(key uint64) error {
	keyBytes := helper.U64ToBytes(key)
	dataBytes := helper.U32ToBytes(d.cursor)
	for _, r := range d.loopRecord {
		dataBytes = append(dataBytes, helper.U64ToBytes(uint64(r))...)
	}
	return db.Put(append([]byte("gypsum-limiters-duration-"), keyBytes...), dataBytes, nil)
}

type unitTimeLimiter struct {
	timeKeyFunc func() uint32
	maxUsage    uint32
	timeMask    uint32
	usage       uint32
	lock        sync.Mutex
}

func NewUnitTimeLimiter(unit string, maxUsage uint32) (*unitTimeLimiter, error) {
	u := unitTimeLimiter{
		maxUsage: maxUsage,
	}
	switch unit {
	case "month":
		u.timeKeyFunc = nowMonth
	case "week":
		u.timeKeyFunc = nowWeek
	case "day":
		u.timeKeyFunc = nowDay
	case "hour":
		u.timeKeyFunc = nowHour
	default:
		return nil, errors.New("unknown unit: " + unit)
	}
	return &u, nil
}

func (u *unitTimeLimiter) SaveToDB(key uint64) error {
	keyBytes := helper.U64ToBytes(key)
	dataBytes := append(helper.U32ToBytes(u.timeMask), helper.U32ToBytes(u.usage)...)
	return db.Put(append([]byte("gypsum-limiters-unit-"), keyBytes...), dataBytes, nil)
}

func (u *unitTimeLimiter) Require() bool {
	u.lock.Lock()
	defer u.lock.Unlock()
	now := u.timeKeyFunc()
	if now == u.timeMask {
		if u.usage < u.maxUsage {
			return false
		} else {
			u.usage += 1
			return true
		}
	} else {
		u.timeMask = now
		u.usage = 1
		return true
	}
}

func (d *LimiterDescriptor) SetDBKey(k []byte) {
	d.databaseKey = k
}

func (d *LimiterDescriptor) ToRule() zero.Rule {
	limiter, err := d.ToLimiter()
	if err != nil {
		log.Error("convert limiter to rule: ", err)
		return RuleAlwaysFalse
	}
	return func(_ *zero.Ctx) bool {
		available := limiter.Require()
		return available
	}
}

func (d *LimiterDescriptor) ToLimiter() (Limiter, error) {
	switch d.LimiterType {
	case UnitTimeLimiterType:
		return NewUnitTimeLimiter(d.Unit, d.MaxUsage)
	case DurationLimiterType:
		return NewDurationLimiter(d.DurationSecond, d.MaxUsage), nil
	default:
		return nil, errors.New(fmt.Sprintf("unknown limiter type: %s", d.LimiterType))
	}
}

// year: ~5, last 4 bits,  take 20~24, 0x00f00000
// week: 1~54,    6 bits,  take 14~19, 0x000fc000
// month: 1~12,   4 bits,  take 10~13, 0x00003c00
// day: 1~31,     5 bits,  take 5~9,   0x000003e0
// hour: 0~23,    5 bits,  take 0~4,   0x0000001f

func nowMonth() uint32 {
	y, m, _ := time.Now().Date()
	return uint32(((y & 0xf) << 20) | (int(m) << 10))
}

func nowDay() uint32 {
	y, m, d := time.Now().Date()
	return uint32(((y & 0xf) << 20) | (int(m) << 10) | (d << 5))
}

func nowWeek() uint32 {
	y, w := time.Now().ISOWeek()
	return uint32(((y & 0xf) << 20) | (w << 14))
}

func nowHour() uint32 {
	now := time.Now()
	y, m, d := now.Date()
	h := now.Hour()
	return uint32(((y & 0xf) << 20) | (int(m) << 10) | (d << 5) | h)
}
