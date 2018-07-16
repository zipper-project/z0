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
	"bufio"
	"errors"
	"fmt"
	"os"
	"reflect"
	"unicode"

	"github.com/ethereum/go-ethereum/log"
	"github.com/naoina/toml"
	"github.com/zipper-project/z0/config"
	"github.com/zipper-project/z0/node"
	"github.com/zipper-project/z0/params"
	"github.com/zipper-project/z0/zcnd"
)

// These settings ensure that TOML keys use the same names as Go struct fields.
var tomlSettings = toml.Config{
	NormFieldName: func(rt reflect.Type, key string) string {
		return key
	},
	FieldToKey: func(rt reflect.Type, field string) string {
		return field
	},
	MissingField: func(rt reflect.Type, field string) error {
		link := ""
		if unicode.IsUpper(rune(rt.Name()[0])) && rt.PkgPath() != "main" {
			link = fmt.Sprintf(", see https://godoc.org/%s#%s for available fields", rt.PkgPath(), rt.Name())
		}
		return fmt.Errorf("field '%s' is not defined in %s%s", field, rt.String(), link)
	},
}

// AllCfg all configs
var AllCfg []Config

//Config  all configs interface
type Config interface {
	Setup()
}

var (
	// log config

	logConfig = new(config.LogConfig)

	// node Config
	nodeConfig = &node.Config{
		Name:   params.ClientIdentifier,
		Logger: log.New(),
	}

	// zcnd Config
	zcndConfig = new(zcnd.Config)

	//z0 Config
	z0_Config = new(z0Config)
)

func init() {
	AllCfg = append(AllCfg, logConfig)
}

func setUpConfig() {
	for _, c := range AllCfg {
		c.Setup()
	}
}

func loadConfig(file string, cfg *z0Config) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	err = tomlSettings.NewDecoder(bufio.NewReader(f)).Decode(cfg)
	// Add file name to errors that have a line number.
	if _, ok := err.(*toml.LineError); ok {
		err = errors.New(file + ", " + err.Error())
	}
	return err
}
