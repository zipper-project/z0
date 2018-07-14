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

package main

import (
	"github.com/zipper-project/z0/config"
)

// AllCfg all configs
var AllCfg []Config

//Config  all configs interface
type Config interface {
	Setup()
}

// log config
var logConfig = new(config.LogConfig)

func init() {
	AllCfg = append(AllCfg, logConfig)
}

func setUpConfig() {
	for _, c := range AllCfg {
		c.Setup()
	}
}
