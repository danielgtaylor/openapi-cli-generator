package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

var (
	// These colors match those in formatter.go
	cReset     = 0
	cLightGray = 247
	cBlue      = 74
	cOrange    = 172
	cRed       = 204
	cGreen     = 150
	cPurple    = 98
	cYellow    = 222
)

var consoleBufPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 100))
	},
}

// ConsoleWriter reads a JSON object per write operation and outputs an
// optionally colored human readable version on the Out writer.
// This has been modified from the ConsoleWriter that ships with zerolog.
type ConsoleWriter struct {
	Out     io.Writer
	NoColor bool
}

func (w ConsoleWriter) Write(p []byte) (n int, err error) {
	var event map[string]interface{}
	d := json.NewDecoder(bytes.NewReader(p))
	d.UseNumber()
	err = d.Decode(&event)
	if err != nil {
		return
	}
	buf := consoleBufPool.Get().(*bytes.Buffer)
	defer consoleBufPool.Put(buf)
	lvlColor := cReset
	level := "????"
	if l, ok := event[zerolog.LevelFieldName].(string); ok {
		if !w.NoColor {
			lvlColor = levelColor(l)
		}
		level = strings.ToUpper(l)
	}
	fmt.Fprintf(buf, "%s %s %s",
		colorize(level, lvlColor, !w.NoColor),
		colorize(path.Base(event[zerolog.CallerFieldName].(string)), cReset, !w.NoColor),
		colorize(event[zerolog.MessageFieldName], cReset, !w.NoColor))
	fields := make([]string, 0, len(event))
	errorField := ""
	for field := range event {
		switch field {
		case zerolog.LevelFieldName, zerolog.TimestampFieldName, zerolog.MessageFieldName, zerolog.CallerFieldName:
			continue
		case "error":
			errorField = fmt.Sprintf("%v\n", event[field])
			continue
		}
		fields = append(fields, field)
	}
	sort.Strings(fields)
	for _, field := range fields {
		fmt.Fprintf(buf, " %s=", colorize(field, lvlColor, !w.NoColor))
		switch value := event[field].(type) {
		case string:
			if needsQuote(value) {
				buf.WriteString(strconv.Quote(value))
			} else {
				buf.WriteString(value)
			}
		case json.Number:
			fmt.Fprint(buf, value)
		default:
			b, err := json.Marshal(value)
			if err != nil {
				fmt.Fprintf(buf, "[error: %v]", err)
			} else {
				fmt.Fprint(buf, string(b))
			}
		}
	}
	buf.WriteByte('\n')
	if errorField != "" {
		buf.Write([]byte(errorField))
	}
	buf.WriteTo(w.Out)
	n = len(p)
	return
}

func colorize(s interface{}, color int, enabled bool) string {
	if !enabled {
		return fmt.Sprintf("%v", s)
	}
	if color == 0 {
		return fmt.Sprintf("\x1b[%dm%v\x1b[0m", color, s)
	}
	return fmt.Sprintf("\x1b[38;5;%dm%v\x1b[0m", color, s)
}

func levelColor(level string) int {
	switch level {
	case "debug":
		return cLightGray
	case "info":
		return cGreen
	case "warn":
		return cOrange
	case "error", "fatal", "panic":
		return cRed
	default:
		return cReset
	}
}

func needsQuote(s string) bool {
	for i := range s {
		if s[i] < 0x20 || s[i] > 0x7e || s[i] == ' ' || s[i] == '\\' || s[i] == '"' {
			return true
		}
	}
	return false
}
