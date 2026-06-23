package serverlog

import (
	"runtime"
	"sync"
)

type frame struct {
	File string
	Line int
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
