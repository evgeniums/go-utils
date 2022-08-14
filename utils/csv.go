package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/dimchansky/utfbom"
)

type Csv struct {
	Delim      byte
	FloatPoint string
	Header     []string
	Lines      [][]string
}

func NewCsv(delim byte, floatPoint string) *Csv {
	c := &Csv{}
	c.Delim = delim
	c.FloatPoint = floatPoint
	return c
}

func DefaultCsv() *Csv {
	return NewCsv(';', ",")
}

func (c *Csv) Content() []byte {

	// prepend BOM
	content := make([]byte, 3)
	content[0] = 0xEF
	content[1] = 0xBB
	content[2] = 0xBF

	// construct header
	for i, field := range c.Header {
		if i != 0 {
			content = append(content, c.Delim)
		}
		content = append(content, field...)
	}
	content = append(content, "\n"...)

	// append strings
	for _, line := range c.Lines {
		for i, field := range line {
			if i != 0 {
				content = append(content, c.Delim)
			}
			content = append(content, field...)
		}
		content = append(content, "\n"...)
	}

	// done
	return content
}

func (c *Csv) FormatFloat(val float64) string {
	str := FloatToStr(val)
	if c.FloatPoint != "." {
		str = strings.ReplaceAll(str, ".", c.FloatPoint)
	}
	return str
}

func (c *Csv) AddLine(fields ...interface{}) {

	line := make([]string, len(fields))

	var str string
	for i, field := range fields {
		switch v := field.(type) {
		case float64:
			str = c.FormatFloat(float64(v))
		default:
			str = fmt.Sprintf("%v", field)
		}
		line[i] = str
	}

	c.Lines = append(c.Lines, line)
}

func (c *Csv) AddLineStrings(fields ...string) {

	line := append([]string{}, fields...)
	c.Lines = append(c.Lines, line)
}

func (c *Csv) SetHeader(fields ...string) {
	c.Header = make([]string, len(fields))
	for i, field := range fields {
		c.Header[i] = field
	}
}

func RemoveBom(str string) string {
	noBom, err := ioutil.ReadAll(utfbom.SkipOnly(bytes.NewReader([]byte(str))))
	if err != nil {
		return str
	}
	return string(noBom)
}
