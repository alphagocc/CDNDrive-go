package main

import (
	"CDNDrive/drivers"
	"CDNDrive/encoders"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/gookit/color"
)

//CDNDrive的格式
type metaJSON struct {
	Time       int64            `json:"time"`
	FileName   string           `json:"filename"`
	Size       int64            `json:"size"`
	Sha1       string           `json:"sha1"`
	BlockDicts []metaJSON_Block `json:"block"`
}

type metaJSON_Block struct {
	URL  string `json:"url"`
	Size int    `json:"size"`
	Sha1 string `json:"sha1"`

	i      int //第几块
	offset int64
}

var _drivers = make([]drivers.Driver, 0)
var _debug bool

func loadDrivers() {
	_drivers = append(_drivers, drivers.NewDriverBilibili())
	_drivers = append(_drivers, drivers.NewDriverBaijia())
	_drivers = append(_drivers, drivers.NewDriverSogou())
}

// metaURL -> Driver
func queryDriverByMetaLink(metaURL string) drivers.Driver {
	for _, d := range _drivers {
		//Meta2Real成功就说明是对应driver的链接
		if d.Meta2Real(metaURL) != "" {
			return d
		}
	}
	return nil
}

func queryDriverByName(name string) drivers.Driver {
	for _, d := range _drivers {
		if name == d.Name() {
			return d
		}
	}
	return nil
}

//resp.Body -> []byte
func readPhotoBytes(r io.Reader, e encoders.Encoder) ([]byte, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return e.Decode(b)
}

type colorLogger_t struct {
	logWriter io.Writer
	prefix    func() string
}

var colorLogger *colorLogger_t

func (p *colorLogger_t) Println(a ...interface{}) {
	header := interface{}(p.prefix())
	b := append([]interface{}{header}, a...)
	color.Println(b...)
}

func (p *colorLogger_t) Printf(f string, a ...interface{}) {
	color.Printf(p.prefix()+" "+f, a...)
}

//下面是抄来的

const (
	// B byte
	B = (int64)(1 << (10 * iota))
	// KB kilobyte
	KB
	// MB megabyte
	MB
	// GB gigabyte
	GB
	// TB terabyte
	TB
	// PB petabyte
	PB
)

// ConvertFileSize 文件大小格式化输出
func ConvertFileSize(size int64, precision ...int) string {
	pint := "6"
	if len(precision) == 1 {
		pint = fmt.Sprint(precision[0])
	}
	if size < 0 {
		return "0B"
	}
	if size < KB {
		return fmt.Sprintf("%dB", size)
	}
	if size < MB {
		return fmt.Sprintf("%."+pint+"fKB", float64(size)/float64(KB))
	}
	if size < GB {
		return fmt.Sprintf("%."+pint+"fMB", float64(size)/float64(MB))
	}
	if size < TB {
		return fmt.Sprintf("%."+pint+"fGB", float64(size)/float64(GB))
	}
	if size < PB {
		return fmt.Sprintf("%."+pint+"fTB", float64(size)/float64(TB))
	}
	return fmt.Sprintf("%."+pint+"fPB", float64(size)/float64(PB))
}

// FormatTime 将 Unix 时间戳, 转换为字符串
func FormatTime(t int64) string {
	tt := time.Unix(t, 0).Local()
	year, mon, day := tt.Date()
	hour, min, sec := tt.Clock()
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d", year, mon, day, hour, min, sec)
}