package serverlog

import (
	"testing"
)

func TestNewLevelSwitch(t *testing.T) {
	ls := NewLevelSwitch()
	if ls == nil {
		t.Fatal("NewLevelSwitch() returned nil")
	}
	for _, lv := range AllLevels {
		if ls.IsEnabled(lv) {
			t.Errorf("new LevelSwitch should not have %v enabled", lv)
		}
	}
}

func TestLevelSwitch_EnableSingleLevel(t *testing.T) {
	ls := NewLevelSwitch()
	ls.EnableSingleLevel(InfoLevel)

	if !ls.IsEnabled(InfoLevel) {
		t.Error("InfoLevel should be enabled")
	}
	// Other levels should still be disabled
	if ls.IsEnabled(ErrorLevel) {
		t.Error("ErrorLevel should not be enabled")
	}
	if ls.IsEnabled(DebugLevel) {
		t.Error("DebugLevel should not be enabled")
	}
}

func TestLevelSwitch_EnableLevels(t *testing.T) {
	ls := NewLevelSwitch()
	ls.EnableLevels(InfoLevel, ErrorLevel, DebugLevel)

	if !ls.IsEnabled(InfoLevel) {
		t.Error("InfoLevel should be enabled")
	}
	if !ls.IsEnabled(ErrorLevel) {
		t.Error("ErrorLevel should be enabled")
	}
	if !ls.IsEnabled(DebugLevel) {
		t.Error("DebugLevel should be enabled")
	}
	if ls.IsEnabled(WarnLevel) {
		t.Error("WarnLevel should not be enabled")
	}
}

func TestLevelSwitch_DisableSingleLevel(t *testing.T) {
	ls := NewLevelSwitch()
	ls.EnableLevels(InfoLevel, ErrorLevel)
	if !ls.IsEnabled(InfoLevel) {
		t.Fatal("InfoLevel should be enabled")
	}

	ls.DisableSingleLevel(InfoLevel)
	if ls.IsEnabled(InfoLevel) {
		t.Error("InfoLevel should be disabled after DisableSingleLevel")
	}
	if !ls.IsEnabled(ErrorLevel) {
		t.Error("ErrorLevel should still be enabled")
	}
}

func TestLevelSwitch_Reset(t *testing.T) {
	ls := NewLevelSwitch()
	ls.EnableLevels(InfoLevel, ErrorLevel, DebugLevel)
	ls.Reset()

	for _, lv := range AllLevels {
		if ls.IsEnabled(lv) {
			t.Errorf("after Reset, %v should not be enabled", lv)
		}
	}
}

func TestLevelSwitch_EnableLevelLte(t *testing.T) {
	ls := NewLevelSwitch()
	// EnableLevelLte(ErrorLevel) should enable Panic, Fatal, Error
	ls.EnableLevelLte(ErrorLevel)

	// ErrorLevel and "lower" (more severe) should be enabled
	if !ls.IsEnabled(PanicLevel) {
		t.Error("PanicLevel should be enabled")
	}
	if !ls.IsEnabled(FatalLevel) {
		t.Error("FatalLevel should be enabled")
	}
	if !ls.IsEnabled(ErrorLevel) {
		t.Error("ErrorLevel should be enabled")
	}
	// "Higher" levels should not be enabled
	if ls.IsEnabled(WarnLevel) {
		t.Error("WarnLevel should not be enabled")
	}
	if ls.IsEnabled(InfoLevel) {
		t.Error("InfoLevel should not be enabled")
	}
	if ls.IsEnabled(DebugLevel) {
		t.Error("DebugLevel should not be enabled")
	}
	if ls.IsEnabled(TraceLevel) {
		t.Error("TraceLevel should not be enabled")
	}
}

func TestLevelSwitchLte(t *testing.T) {
	ls := LevelSwitchLte(ErrorLevel)
	if !ls.IsEnabled(ErrorLevel) {
		t.Error("ErrorLevel should be enabled")
	}
	if !ls.IsEnabled(PanicLevel) {
		t.Error("PanicLevel should be enabled")
	}
	if ls.IsEnabled(InfoLevel) {
		t.Error("InfoLevel should not be enabled")
	}
}
