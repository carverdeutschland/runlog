package runlog

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"time"
)

type rotatewriter struct {
	filename   string
	rotateTime time.Time // 下次转换的时间
	rotateSize int64
	out        io.WriteCloser //
	maxage     int32
	dayIndex   int32
	curSize    int64
	histroyNum int32
	zip        bool
}

func getRotateTime() time.Time {
	now := time.Now()

	t1, _ := time.Parse("2006-01-02 15:04:05", now.Format("2006-01-02 15:04:05"))
	t2, _ := time.Parse("2006-01-02 15:04:05", now.UTC().Format("2006-01-02 15:04:05"))

	// 我们需要在当地时间进行截断，所以要加上这个差值
	nowt := now.Add(t1.Sub(t2)).Truncate(24 * time.Hour)

	//
	return nowt.Add(t2.Sub(t1)).Add(24 * time.Hour)
}

func (r *rotatewriter) Init(f string) error {
	if f != "" {
		r.filename = f
	}

	fp, err := os.OpenFile(r.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	fifo, err := fp.Stat()
	if err == nil {
		r.curSize = fifo.Size()
	}
	r.out = fp

	r.rotateTime = getRotateTime()
	return nil
}

// reason:0  时间到了，1:文件大小已满
func (r *rotatewriter) rotate(reason int) error {
	err := r.out.Close()

	newName := r.filename + "." +
		r.rotateTime.Add(-24*time.Hour).Format("2006-01-02") +
		fmt.Sprintf(".%02d", r.dayIndex)

	err = os.Rename(r.filename, newName)
	if err != nil {
		return err
	}

	fp, err := os.OpenFile(r.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	if reason == 0 {
		r.rotateTime = r.rotateTime.Add(24 * time.Hour)
		r.dayIndex = 0
	} else {
		r.dayIndex++
	}
	r.out = fp
	r.curSize = 0

	if r.zip {
		r.zipFile(newName)
	}
	go r.purge()

	return nil
}

func (r *rotatewriter) purge() {
	idx := strings.LastIndex(r.filename, "/")
	dir := r.filename[:idx]
	fn := r.filename[idx+1:]
	fileInfoList, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Println("purge failed, 1:", err.Error())
		return
	}

	fileLeft := make([]string, 0, 20)
	t := r.rotateTime.Add(-24 * time.Duration(r.maxage+2) * time.Hour)
	toRmBefore := fn + "." + t.Format("2006-01-02")
	for i := range fileInfoList {
		if fileInfoList[i].IsDir() {
			continue
		}

		// 当前正在用的日志文件不能删掉
		if strings.HasSuffix(fileInfoList[i].Name(), ".log") {
			continue
		}

		if strings.Compare(toRmBefore, fileInfoList[i].Name()) >= 0 {
			sfile := path.Join(dir, fileInfoList[i].Name())

			// 日志目录下可能有其他文件，如果不是这个组件管理的日志文件，就忽略，不能删除
			if !strings.HasPrefix(sfile, r.filename) {
				continue
			}

			err := os.Remove(sfile)
			if err != nil {
				fmt.Println("purge failed:", err.Error())
			} else {
				fmt.Println("purged:", sfile)
			}
		} else {
			fileLeft = append(fileLeft, fileInfoList[i].Name())
		}
	}

	if r.histroyNum > 0 && len(fileLeft) > int(r.histroyNum) {
		sort.Strings(fileLeft)
		for i := 0; i < len(fileLeft)-int(r.histroyNum); i++ {
			err := os.Remove(path.Join(dir, fileLeft[i]))
			if err != nil {
				fmt.Println(err.Error())
			}
		}
	}
}

func (r *rotatewriter) Write(p []byte) (n int, err error) {
	now := time.Now()
	if now.After(r.rotateTime) {
		err := r.rotate(0)
		if err != nil {
			fmt.Println("rotate failed:", err.Error())
		}
	}

	if r.curSize > r.rotateSize {
		err := r.rotate(1)
		if err != nil {
			fmt.Println("rotate failed:", err.Error())
		}
	}

	r.curSize += int64(len(p))

	return r.out.Write(p)
}

func (r *rotatewriter) zipFile(flog string) error {
	zipfile, err := os.Create(flog + ".zip")
	if err != nil {
		return err
	}
	defer zipfile.Close()

	zw := zip.NewWriter(zipfile)
	defer zw.Close()

	info, err := os.Stat(flog)
	if err != nil {
		return err
	}

	// 获取压缩头信息
	head, err := zip.FileInfoHeader(info)

	if err != nil {
		return err
	}

	// 指定文件压缩方式 默认为 Store 方式 该方式不压缩文件 只是转换为zip保存
	head.Method = zip.Deflate

	fw, err := zw.CreateHeader(head)
	if err != nil {
		return err
	}

	file, err := os.Open(flog)
	if err != nil {
		return err
	}

	// 写入文件到压缩包中
	_, err = io.Copy(fw, file)
	if err != nil {
		return err
	}

	file.Close()
	os.Remove(flog)

	return nil
}
