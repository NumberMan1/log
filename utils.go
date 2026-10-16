package serverlog

import (
	"runtime"
	"strings"
	"sync"
)

type frame struct {
	File string
	Line int
}

// Resolve caller frames by package, independent of wrapper depth.
func applicationFrames() []*frame {
	var pcs [32]uintptr
	n := runtime.Callers(2, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	var result []*frame
	for {
		f, more := frames.Next()
		if (!strings.HasPrefix(f.Function, "github.com/NumberMan1/log.") || strings.HasSuffix(f.File, "_test.go")) && !strings.HasPrefix(f.Function, "runtime.") {
			result = append(result, &frame{File: f.File, Line: f.Line})
		}
		if !more {
			return result
		}
	}
}

func applicationFrame() *frame {
	if frames := applicationFrames(); len(frames) > 0 {
		return frames[0]
	}
	return &frame{}
}

// skip 从1开始，每增加一层就应该加1
func stackTraceFrame(skip int) *frame {
	_, f, line, ok := runtime.Caller(skip)
	if !ok {
		return &frame{}
	}

	return &frame{File: f, Line: line}
}

var pcPool = sync.Pool{
	New: func() interface{} {
		ptr := make([]uintptr, 10)
		return &ptr
	},
}

// skip 从1开始，每增加一层就应该加1
func stackTraceFrames(skip int) (result []*frame) {
	pcs := pcPool.Get().(*[]uintptr)
	defer pcPool.Put(pcs)

	waitReadCount := runtime.Callers(skip, *pcs)
	if waitReadCount == 0 {
		return
	}
	frames := runtime.CallersFrames(*pcs)
	for {
		waitReadCount--

		f, more := frames.Next()
		result = append(result, &frame{File: f.File, Line: f.Line})
		if !more || waitReadCount <= 0 {
			break
		}
	}
	return
}
