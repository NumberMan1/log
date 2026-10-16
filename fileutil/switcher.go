package fileutil

import (
	"time"
)

var (
	// DAY_SWITCHER 按本地日历日期命名，后续轮转间隔为 24 小时。
	DAY_SWITCHER = daySwitch{}
	// HOURS_SWITCHER 按本地小时命名，后续轮转间隔为 1 小时。
	HOURS_SWITCHER = hoursSwitch{}
)

// Switcher 决定首次和后续轮转间隔，以及不含序号和扩展名的目录、文件名。
// 路径计算可能在构造阶段执行；实现应返回正间隔并支持其使用方式要求的并发访问。
type Switcher interface {
	// FirstSwitchTime 返回首次轮转前等待的正时长。
	FirstSwitchTime() time.Duration
	// NextSwitchTime 返回每次后续轮转前等待的正时长。
	NextSwitchTime() time.Duration
	// DirAndFileName 返回目录与文件名基础路径，不含序号及扩展名。
	DirAndFileName(base string) (dir, file string)
}

func now() time.Time {
	return time.Now()
}

type daySwitch struct{}

// FirstSwitchTime 返回从当前时间到首次日切边界的间隔。
func (_ daySwitch) FirstSwitchTime() time.Duration {
	// 到明天凌晨间隔多长时间
	now := now()
	tt := time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, now.Location(),
	).Add(time.Hour * 24).Unix()
	return time.Second * (time.Duration(tt - now.Unix()))
}

// NextSwitchTime 返回固定 24 小时的后续间隔。
func (_ daySwitch) NextSwitchTime() time.Duration {
	return 24 * time.Hour
}

// DirAndFileName 按本地日期生成目录与日文件名。
func (_ daySwitch) DirAndFileName(base string) (dir, file string) {
	now := now()
	dir = base + "/" + now.Format("2006/01/02/")
	file = dir + now.Format("2006-01-02")
	return
}

type hoursSwitch struct{}

// FirstSwitchTime 返回到下个本地整点的间隔。
func (_ hoursSwitch) FirstSwitchTime() time.Duration {
	// 到下一个整点间隔多长时间
	now := now()
	tt := time.Date(
		now.Year(), now.Month(), now.Day(),
		now.Hour(), 0, 0, 0, now.Location(),
	).Add(time.Hour).Unix()
	return time.Second * (time.Duration(tt - now.Unix()))
}

// NextSwitchTime 返回固定一小时的后续间隔。
func (_ hoursSwitch) NextSwitchTime() time.Duration {
	return time.Hour
}

// DirAndFileName 按本地日期生成目录与小时文件名。
func (_ hoursSwitch) DirAndFileName(base string) (dir, file string) {
	now := now()
	dir = base + "/" + now.Format("2006/01/02/")
	file = dir + now.Format("2006-01-02_15")
	return
}
