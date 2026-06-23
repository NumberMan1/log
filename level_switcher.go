package serverlog

import "sync/atomic"

type LevelSwitch struct {
	enabledLevel uint32
}

func LevelSwitchLte(level Level) *LevelSwitch {
	return NewLevelSwitch().EnableLevelLte(level)
}

func NewLevelSwitch() *LevelSwitch {
	return &LevelSwitch{}
}

func (ls *LevelSwitch) IsEnabled(level Level) bool {
	lvBit := 1 << level
	lv := atomic.LoadUint32(&ls.enabledLevel)
	return (lv & uint32(lvBit)) > 0
}

func (ls *LevelSwitch) EnableLevels(levels ...Level) *LevelSwitch {
	for _, lv := range levels {
		ls.EnableSingleLevel(lv)
	}

	return ls
}

func (ls *LevelSwitch) EnableSingleLevel(level Level) *LevelSwitch {
	lvBit := 1 << level
	atomic.StoreUint32(&ls.enabledLevel, ls.enabledLevel|uint32(lvBit))

	return ls
}

func (ls *LevelSwitch) DisableSingleLevel(level Level) *LevelSwitch {
	lvBit := 1 << level
	atomic.StoreUint32(&ls.enabledLevel, ls.enabledLevel^uint32(lvBit))

	return ls
}

func (ls *LevelSwitch) Reset() *LevelSwitch {
	atomic.StoreUint32(&ls.enabledLevel, 0)

	return ls
}

// trace > debug > info > event > warn > error > fatal > panic
func (ls *LevelSwitch) EnableLevelLte(level Level) *LevelSwitch {
	lvBit := uint32(1)
	for i := 0; i < int(level); i++ {
		lvBit <<= 1
		lvBit |= 1
	}

	atomic.StoreUint32(&ls.enabledLevel, ls.enabledLevel|lvBit)

	return ls
}
