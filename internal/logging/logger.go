package logging

import (
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog"
)

func NewLogger(w io.Writer, level zerolog.Level) zerolog.Logger {
	return zerolog.New(zerolog.ConsoleWriter{
		Out:        w,
		TimeFormat: "2006-01-02 15:04:05.000",
		FormatLevel: func(i any) string {
			return strings.ToUpper(fmt.Sprintf("[%-5s]", i))
		},
		FormatFieldName: func(i any) string {
			return fmt.Sprintf("\n\t%s:", i)
		},
		NoColor: true,
	}).Level(level).With().Timestamp().Logger()
}
