package sender

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"regexp"
	"runtime"
	"strings"

	"github.com/pkg/errors"
)

var (
	emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	phoneRegexp = regexp.MustCompile("^\\+[0-9]+[\\s]+[\\s0-9]+$")
)

type EmailAddr string

func NewEmail(emailAddr string) (EmailAddr, error) {
	if !emailRegexp.MatchString(emailAddr) {
		return "", Error(errors.New("Invalid email format"))
	}
	return EmailAddr(strings.ToLower(emailAddr)), nil
}

func (e EmailAddr) String() string {
	return string(e)
}

type Phone struct {
	Area      string
	SubNumber string
}

func NewPhone(phone string) (*Phone, error) {
	if !phoneRegexp.MatchString(phone) {
		return nil, Error(errors.New("Invalid phone format"))
	}

	arr := strings.Split(
		strings.Replace(
			strings.Trim(phone, " "),
			" ",
			"-",
			1,
		),
		"-")
	if len(arr) < 2 {
		return nil, errors.New("Invalid phone format")
	}
	return &Phone{arr[0], strings.Replace(arr[1], " ", "", -1)}, nil
}

func (p Phone) String() string {
	return fmt.Sprintf("%s %s", p.Area, p.SubNumber)
}

func MD5(ctx string) string {
	h := md5.New()
	h.Write([]byte(ctx))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func Error(err error) error {
	var (
		pc, file, line, _ = runtime.Caller(1)
		funcname          = func(name string) string {
			i := strings.LastIndex(name, "/")
			name = name[i+1:]
			i = strings.Index(name, ".")
			return name[i+1:]
		}
		callName = funcname(runtime.FuncForPC(pc).Name())
	)
	return errors.Wrapf(err, "%s:%s:%d", callName, file, line)
}
