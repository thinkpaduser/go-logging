package logging

import (
	"encoding/json"
	"fmt"
	"github.com/hhkbp2/go-strftime"
	"regexp"
)

const (
	JSONPrefix = ""
	JSONIdent = "  "
)

type JSONFormatter struct {
	fields map[string]Formatter
	pretty bool
}

// Initialize the formatter with specified format strings.
// Allow for specialized date formatting with the dateFormat argument.
func NewJSONFormatter(formatFields map[string]string, dateFormat string, pretty bool) Formatter {
	logFields := make(map[string]Formatter)
	toFormatTime := false
	size := 0
	keywordRe := `\(([_\w]+)\)`
	keywordRegexp := regexp.MustCompile(keywordRe)
	for k, v := range formatFields {
		var replaceFuncs []RecordValueReceiver
		f1 := func(match string) string {
			if match == "%%" {
				return "%"
			}
			if match == "%(asctime)s" {
				toFormatTime = true
			}
			size++
			matches := keywordRegexp.FindStringSubmatch(match)
			keyword := matches[1]

			valueReceiver := func(fieldName string) RecordValueReceiver {
				var extractFunc RecordValueReceiver
				extractFunc = func(record *LogRecord) interface{} {
					return GetValueForField(fieldName, record)
				}
				return extractFunc
			}(keyword)
			replaceFuncs = append(replaceFuncs, valueReceiver)

			if keyword == "levelno" {
				return "%d"
			}
			return "%s"
		}

		strFormat := formatRe.ReplaceAllStringFunc(v, f1)
		getFormatArgsFunc := func(record *LogRecord) []interface{} {
			result := make([]interface{}, 0, len(replaceFuncs))
			for _, f := range replaceFuncs {
				result = append(result, f(record))
			}
			return result
		}
		var dateFormatter *strftime.Formatter
		if toFormatTime {
			dateFormatter = strftime.NewFormatter(dateFormat)
		}
		standartFormatter := &StandardFormatter{
			format:            v,
			strFormat:         strFormat,
			getFormatArgsFunc: getFormatArgsFunc,
			toFormatTime:      toFormatTime,
			dateFormat:        dateFormat,
			dateFormatter:     dateFormatter,
		}
		logFields[k] = standartFormatter
	}

	JSONFormat := &JSONFormatter{
		fields: logFields,
		pretty: pretty,
	}
	return JSONFormat
}

func ( self *JSONFormatter ) Format( record *LogRecord ) string {
	record.GetMessage()
	serializeFields := make(map[string]string)
	for k, v := range self.fields {
		readyValue := v.Format(record)
		serializeFields[k] = readyValue
	}

	var (
		stringToLog string
		bytesToLog []byte
		err error
	)
	if self.pretty {
		bytesToLog, err = json.MarshalIndent(serializeFields, JSONPrefix, JSONIdent)
	} else {
		bytesToLog, err = json.Marshal(serializeFields)
	}

	if err != nil {
		stringToLog = fmt.Sprintf("[error:%s]", err)
	} else {
		stringToLog = string(bytesToLog)
	}
	return stringToLog
}
