package runlog

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

// 日志输出级别：0:fatal 1:error 2:warn 3:info 4:debug 5:trace
type DebugLevelCtrl int32

const (
	DbgFatal DebugLevelCtrl = 0
	DbgError DebugLevelCtrl = 1
	DbgWarn  DebugLevelCtrl = 2
	DbgInfo  DebugLevelCtrl = 3
	DbgDebug DebugLevelCtrl = 4
	DbgTrace DebugLevelCtrl = 5
)

// debug configuration
type DebugConf struct {
	DbgLogFile   string            // 日志文件输出路径
	DbgLevel     DebugLevelCtrl    // 输出日志最低级别，低于这个级别的不会输出
	WithFunc     bool              // 日志末尾自动输出函数
	WithFileLine bool              // 日志末尾输出所在行
	WithLongFile bool              // 是否全文件名
	WhiteTable   bool              // true:使用白名单，否则使用黑名单
	Include      []string          //
	includemap   map[string]string // 白名单列表
	BlackTable   bool              // 使用黑名单
	Exclude      []string
	excludemap   map[string]string // 黑名单列表
	MaxAge       int32             // 最长保留几天
	RotationSize int64             // 日志切割大小，单位k
	HistroyNum   int32             // 被压缩的日志最多保留几个
	Zip          bool              // 是否压缩历史日志
}

var dbgConf *DebugConf
var rwt *rotatewriter

func Init(c *DebugConf) error {
	if len(c.DbgLogFile) < 5 {
		return errors.New("logfile path required")
	}

	// 没有配置，默认7天
	if c.MaxAge <= 0 {
		c.MaxAge = 7
	}

	// 默认100M，这里的单位是k
	if c.RotationSize <= 0 {
		c.RotationSize = 102400
	}

	var writer io.Writer
	if c.DbgLogFile == "stdout" {
		writer = os.Stdout
	} else {
		idx := strings.LastIndex(c.DbgLogFile, "/")
		_ = os.MkdirAll(c.DbgLogFile[:idx], 0755)

		rw := &rotatewriter{
			maxage:     c.MaxAge,
			rotateSize: c.RotationSize * 1024,
			histroyNum: c.HistroyNum,
			zip:        c.Zip,
		}

		err := rw.Init(c.DbgLogFile)
		if err != nil {
			return err
		}

		writer = rw
		rwt = rw
	}

	logflag := 0
	if c.WithFileLine {
		if c.WithLongFile {
			logflag = log.LstdFlags | log.Llongfile
		} else {
			logflag = log.LstdFlags | log.Lshortfile
		}
	} else {
		logflag = log.LstdFlags
	}

	logger = log.New(writer, "", logflag)

	c.slice2map()
	dbgConf = c
	return nil
}

func (c *DebugConf) slice2map() {
	if c.Include != nil {
		c.includemap = make(map[string]string)
		for _, v := range c.Include {
			c.includemap[v] = ""
		}
	}

	if c.Exclude != nil {
		c.excludemap = make(map[string]string)
		for _, v := range c.Exclude {
			c.excludemap[v] = ""
		}
	}
}

func SetNewConf(c *DebugConf) {
	if dbgConf.DbgLevel != c.DbgLevel {
		Warn("dbg level changed %d --> %d", dbgConf.DbgLevel, c.DbgLevel)
	}
	c.slice2map()
	dbgConf = c
}

func GetConf() *DebugConf {
	return dbgConf
}

// debug调试输出，全部输出到标准输出，跑测试用例的时候适用
func SetupDebugRunlog() {
	cf := &DebugConf{
		DbgLogFile:   "stdout",
		DbgLevel:     5,
		WithFunc:     true,
		WithFileLine: true,
	}

	err := Init(cf)
	if err != nil {
		fmt.Println(err.Error())
	}
}
