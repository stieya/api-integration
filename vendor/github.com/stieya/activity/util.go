package activity

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

//ToString convert type (int, int64, float32, float64, byte, and []bytes) to string
//Parameter p is optional and only used in converting float
func ToString(n interface{}, p ...int) string {
	var t string

	switch n.(type) {
	case bool:
		t = strconv.FormatBool(n.(bool))
	case int:
		t = strconv.Itoa(n.(int))
	case int64:
		t = strconv.FormatInt(n.(int64), 10)
	case float32:
		if len(p) > 0 {
			t = strconv.FormatFloat(float64(n.(float32)), 'f', p[0], 64)
		} else {
			t = strconv.FormatFloat(float64(n.(float32)), 'f', -1, 64)
		}
	case float64:
		if len(p) > 0 {
			t = strconv.FormatFloat(n.(float64), 'f', p[0], 64)
		} else {
			t = strconv.FormatFloat(n.(float64), 'f', -1, 64)
		}
	case byte:
		t = string(n.(byte))
	case []byte:
		t = string(n.([]byte))
	case string:
		t = n.(string)
	}

	return t
}

//EncodeMD5 : encrypt to MD5 input string, output to string
func EncodeMD5(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

func NewFileLog(path, name string) (*FileLog, error) {
	fileLog := &FileLog{}

	p := fmt.Sprintf("%s%s", path, name)
	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fileLog, err
	}

	file, err := os.OpenFile(p, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return fileLog, err
	}

	var logging = log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)
	logging.SetOutput(file)

	return &FileLog{
		Conn:     logging,
		FilePath: path,
		FileName: name,
	}, nil
}

func (fl *FileLog) WriteWithID(id string, str interface{}) {
	fl.Conn.Printf("[%s] %q", id, str)
}

func (fl *FileLog) Write(str interface{}) {
	fl.Conn.Printf("%s", str)
}

func (fl *FileLog) WriteWithRandomID(str interface{}) {
	fl.Conn.Printf("[%s] %q", EncodeMD5(ToString(str)), str)
}

func checkRequired(name, str string, option ...string) error {
	if str == "" {
		errorStr := fmt.Sprintf("%s must not be empty", name)
		return fmt.Errorf(errorStr)
	}
	return nil
}

func checkDateTime(name, str string, option ...string) error {
	if str == "" {
		return nil
	}

	_, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		errorStr := fmt.Sprintf("%s not valid", name)
		return fmt.Errorf(errorStr)
	}

	return nil
}

// formatRequest generates ascii representation of a request
func formatRequest(r *http.Request) ([]byte, error) {
	return httputil.DumpRequest(r, true)
}

// formatResponse generates ascii representation of a response
func formatResponse(r *http.Response) ([]byte, error) {
	return httputil.DumpResponse(r, true)
}

func writeLog(logger *FileLog, s string, userID int64) {
	if logger != nil {
		go logger.WriteWithRandomID(fmt.Sprintf("[%d] - %s", userID, s))
	}
}
