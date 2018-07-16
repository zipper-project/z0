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
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var dataDir string

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init <genesisPath>",
	Short: "Bootstrap and initialize a new genesis block",
	Long:  `Bootstrap and initialize a new genesis block`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initGenesis(args); err != nil {
			fmt.Println(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&nodeConfig.DataDir, "datadir", "d", defaultDataDir(), "Data directory for the databases and keystore")
}

// initGenesis will initialise the given JSON format genesis file and writes it as
// the zero'd block (i.e. genesis) or will fail hard if it can't succeed.
func initGenesis(args []string) error {
	// Make sure we have a valid genesis JSON
	genesisPath := args[0]
	if len(genesisPath) == 0 {
		return errors.New("Must supply path to genesis JSON file")
	}
	file, err := os.Open(genesisPath)
	if err != nil {
		return fmt.Errorf("Failed to read genesis file: %v", err)
	}
	defer file.Close()

	return nil
}

// defaultDataDir is the default data directory to use for the databases and other
// persistence requirements.
func defaultDataDir() string {
	// Try to place the data folder in the user's home dir
	home := homeDir()
	if home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "z0_ledger")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", "z0_ledger")
		} else {
			return filepath.Join(home, ".z0_ledger")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}
