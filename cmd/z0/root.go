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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/log"
	"github.com/spf13/cobra"
	"github.com/zipper-project/z0/node"
	"github.com/zipper-project/z0/zcnd"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "z0",
	Short: "z0 is a Leading High-performance Ledger",
	Long:  `z0 is a Leading High-performance Ledger`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		setUpConfig()
		node := makeNode()
		if err := registerService(node); err != nil {
			log.Error("z0 start node failed.", "err", err)
		}

		if err := startNode(node); err != nil {
			log.Error("z0 start node failed.", "err", err)
		}

		node.Wait()
	},
}

func makeNode() *node.Node {
	// load defaults config
	z0_Config.NodeCfg = nodeConfig
	z0_Config.ZcndCfg = zcndConfig

	//  load config file.
	if file := z0_Config.ConfigFileFlag; file != "" {
		if err := loadConfig(file, z0_Config); err != nil {
			log.Error("load config file %v", err)
		}
	}

	return node.New(z0_Config.NodeCfg)
}

// start up the node itself
func startNode(stack *node.Node) error {
	log.Info("z0 start...")
	if err := stack.Start(); err != nil {
		return err
	}
	go func() {
		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sigc)
		<-sigc
		log.Info("Got interrupt, shutting down...")
		go stack.Stop()
		for i := 10; i > 0; i-- {
			<-sigc
			if i > 1 {
				log.Warn("Already shutting down, interrupt more to panic.", "times", i-1)
			}
		}
	}()
	return nil
}

func registerService(stack *node.Node) error {
	var err error
	// register zcnd
	err = stack.Register(func(ctx *node.ServiceContext) (node.Service, error) {
		return zcnd.New(ctx, z0_Config.ZcndCfg)
	})
	return err
}

func init() {
	// todo defining own help
	// RootCmd.SetHelpTemplate( )
	// RootCmd.SetHelpFunc()
	// RootCmd.SetHelpCommand()

	// logging
	RootCmd.Flags().BoolVar(&logConfig.PrintOrigins, "debug", false, "Prepends log messages with call-site location (file and line number)")
	RootCmd.Flags().IntVar(&logConfig.Level, "level", 3, "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail")
	RootCmd.Flags().StringVar(&logConfig.Vmodule, "vmodule", "", "Per-module verbosity: comma-separated list of <pattern>=<level> (e.g. eth/*=5,p2p=4)")
	RootCmd.Flags().StringVar(&logConfig.BacktraceAt, "backtrace", "", "Request a stack trace at a specific logging statement (e.g. \"block.go:271\")")

	// node
	RootCmd.Flags().StringVarP(&nodeConfig.DataDir, "datadir", "d", defaultDataDir(), "Data directory for the databases and keystore")

	// config file
	RootCmd.Flags().StringVarP(&z0_Config.ConfigFileFlag, "config", "c", "", "TOML configuration file")

}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
