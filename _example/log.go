package main

import (
	"os"
	"runtime"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

func init() {
	log = zerolog.New(zerolog.ConsoleWriter{
		Out:     os.Stdout,
		NoColor: runtime.GOOS == "windows",
	}).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
}
