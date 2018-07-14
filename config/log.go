// Copyright 2018 The zipper team Authors
// This file is part of the z0 library.
//
// The z0 library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The z0 library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the z0 library. If not, see <http://www.gnu.org/licenses/>.

package config

import (
	"io"
	"os"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/log/term"
	colorable "github.com/mattn/go-colorable"
)

var glogger *log.GlogHandler

// LogConfig log config
type LogConfig struct {
	PrintOrigins bool
	Level        int
	Vmodule      string
	BacktraceAt  string
}

func init() {
	usecolor := term.IsTty(os.Stderr.Fd()) && os.Getenv("TERM") != "dumb"
	output := io.Writer(os.Stderr)
	if usecolor {
		output = colorable.NewColorableStderr()
	}
	glogger = log.NewGlogHandler(log.StreamHandler(output, log.TerminalFormat(usecolor)))
}

//Setup initializes logging based on the LogConfig
func (lc *LogConfig) Setup() {
	// logging
	log.PrintOrigins(lc.PrintOrigins)
	glogger.Verbosity(log.Lvl(lc.Level))
	glogger.Vmodule(lc.Vmodule)
	glogger.BacktraceAt(lc.BacktraceAt)
	log.Root().SetHandler(glogger)
}
