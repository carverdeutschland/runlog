package runlog

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

var (
	logger *log.Logger
	depth  = 2
)

func getRunFuncName() (string, bool) {
	fName := ""
	rc := true

	pc, _, _, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	fName = f.Name()

	idx := strings.LastIndex(fName, "/")
	fName = fName[idx+1:]

	// 如果使用白名单，白名单里面没有这个函数，那么就不记日志
	if dbgConf.WhiteTable {
		if _, ok := dbgConf.includemap[fName]; !ok {
			rc = false
		}
	} else if dbgConf.BlackTable {
		if _, ok := dbgConf.excludemap[fName]; ok {
			rc = false
		}
	}

	return fName, rc
}

func Trace(format string, args ...interface{}) {
	if dbgConf == nil {
		return
	}
	if dbgConf.DbgLevel < DbgTrace {
		return
	}

	funcName, b := getRunFuncName()
	if !b {
		return
	}

	s := fmt.Sprintf(format, args...)
	if dbgConf.WithFunc {
		_ = logger.Output(depth, fmt.Sprintf("T: %s [%s]", s, funcName))
	} else {
		_ = logger.Output(depth, "T: "+s)
	}
}

func Debug(format string, args ...interface{}) {
	if dbgConf == nil {
		return
	}
	if dbgConf.DbgLevel < DbgDebug {
		return
	}

	funcName, b := getRunFuncName()
	if !b {
		return
	}

	s := fmt.Sprintf(format, args...)
	if dbgConf.WithFunc {
		_ = logger.Output(depth, fmt.Sprintf("D: %s [%s]", s, funcName))
	} else {
		_ = logger.Output(depth, "D: "+s)
	}
}

func Info(format string, args ...interface{}) {
	if dbgConf == nil {
		return
	}
	if dbgConf.DbgLevel < DbgInfo {
		return
	}

	funcName, b := getRunFuncName()
	if !b {
		return
	}

	s := fmt.Sprintf(format, args...)
	if dbgConf.WithFunc {
		_ = logger.Output(depth, fmt.Sprintf("I: %s [%s]", s, funcName))
	} else {
		_ = logger.Output(depth, "I: "+s)
	}
}

func Warn(format string, args ...interface{}) {
	if dbgConf == nil {
		return
	}
	if dbgConf.DbgLevel < DbgWarn {
		return
	}

	funcName, b := getRunFuncName()
	if !b {
		return
	}

	s := fmt.Sprintf(format, args...)
	if dbgConf.WithFunc {
		_ = logger.Output(depth, fmt.Sprintf("W: %s [%s]", s, funcName))
	} else {
		_ = logger.Output(depth, "W: "+s)
	}
}

func Err(format string, args ...interface{}) {
	if dbgConf == nil {
		return
	}
	if dbgConf.DbgLevel < DbgError {
		return
	}

	funcName, b := getRunFuncName()
	if !b {
		return
	}

	s := fmt.Sprintf(format, args...)
	if dbgConf.WithFunc {
		_ = logger.Output(depth, fmt.Sprintf("E: %s [%s]", s, funcName))
	} else {
		_ = logger.Output(depth, "E: "+s)
	}
}

func Fatal(format string, args ...interface{}) {
	if dbgConf == nil {
		return
	}
	funcName, b := getRunFuncName()
	if !b {
		return
	}

	s := fmt.Sprintf(format, args...)
	if dbgConf.WithFunc {
		_ = logger.Output(depth, fmt.Sprintf("F: %s [%s]", s, funcName))
	} else {
		_ = logger.Output(depth, "F: "+s)
	}
}
