package serverlog

import "sync/atomic"

// LevelSwitch 通过原子位掩码管理启用等级，支持并发增删和查询。
type LevelSwitch struct {
	enabledLevel uint32
}

// LevelSwitchLte 创建启用指定等级及更高严重程度等级的开关。
func LevelSwitchLte(level Level) *LevelSwitch {
	return NewLevelSwitch().EnableLevelLte(level)
}

// NewLevelSwitch 创建全部等级关闭的开关。
func NewLevelSwitch() *LevelSwitch {
	return &LevelSwitch{}
}

// IsEnabled 查询等级是否启用；超出已定义范围的等级返回 false。
func (ls *LevelSwitch) IsEnabled(level Level) bool {
	if level > TraceLevel {
		return false
	}
	lvBit := 1 << level
	lv := atomic.LoadUint32(&ls.enabledLevel)
	return (lv & uint32(lvBit)) > 0
}

// EnableLevels 逐个启用指定等级；整个列表不是一次原子事务。
func (ls *LevelSwitch) EnableLevels(levels ...Level) *LevelSwitch {
	for _, lv := range levels {
		ls.EnableSingleLevel(lv)
	}

	return ls
}

// EnableSingleLevel 原子启用单个有效等级，重复调用幂等。
func (ls *LevelSwitch) EnableSingleLevel(level Level) *LevelSwitch {
	if level > TraceLevel {
		return ls
	}
	lvBit := 1 << level
	ls.update(uint32(lvBit), 0)

	return ls
}

// DisableSingleLevel 原子禁用单个有效等级，不会把关闭状态反向开启。
func (ls *LevelSwitch) DisableSingleLevel(level Level) *LevelSwitch {
	if level > TraceLevel {
		return ls
	}
	lvBit := 1 << level
	ls.update(0, uint32(lvBit))

	return ls
}

// Reset 原子关闭所有等级。
func (ls *LevelSwitch) Reset() *LevelSwitch {
	atomic.StoreUint32(&ls.enabledLevel, 0)

	return ls
}

// EnableLevelLte 启用给定等级及更高严重程度的等级，忽略无效值。
func (ls *LevelSwitch) EnableLevelLte(level Level) *LevelSwitch {
	if level > TraceLevel {
		return ls
	}
	lvBit := uint32(1)
	for i := 0; i < int(level); i++ {
		lvBit <<= 1
		lvBit |= 1
	}

	ls.update(lvBit, 0)

	return ls
}

func (ls *LevelSwitch) update(enable, disable uint32) {
	for {
		old := atomic.LoadUint32(&ls.enabledLevel)
		if atomic.CompareAndSwapUint32(&ls.enabledLevel, old, (old|enable)&^disable) {
			return
		}
	}
}
