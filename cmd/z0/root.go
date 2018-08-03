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
	//  load config file.
	if file := zconfig.ConfigFileFlag; file != "" {
		if err := loadConfig(file, zconfig); err != nil {
			log.Error("load config file %v", err)
		}
	}
	return node.New(zconfig.NodeCfg)
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
		return zcnd.New(ctx, zconfig.ZcndCfg)
	})
	return err
}

func init() {
	falgs := RootCmd.Flags()
	// logging
	falgs.BoolVar(&logConfig.PrintOrigins, "log_debug", false, "Prepends log messages with call-site location (file and line number)")
	falgs.IntVar(&logConfig.Level, "log_level", 3, "Logging verbosity: 0=silent, 1=error, 2=warn, 3=info, 4=debug, 5=detail")
	falgs.StringVar(&logConfig.Vmodule, "log_vmodule", "", "Per-module verbosity: comma-separated list of <pattern>=<level> (e.g. eth/*=5,p2p=4)")
	falgs.StringVar(&logConfig.BacktraceAt, "log_backtrace", "", "Request a stack trace at a specific logging statement (e.g. \"block.go:271\")")

	// config file
	falgs.StringVarP(&zconfig.ConfigFileFlag, "config", "c", "", "TOML configuration file")

	// node
	falgs.StringVarP(&zconfig.NodeCfg.DataDir, "datadir", "d", defaultDataDir(), "Data directory for the databases and keystore")

	// zcnd
	falgs.IntVar(&zconfig.ZcndCfg.DatabaseCache, "zcnd_databasecache", zconfig.ZcndCfg.DatabaseCache, "Megabytes of memory allocated to internal database caching")
	falgs.IntVar(&zconfig.ZcndCfg.TrieCache, "zcnd_triecache", zconfig.ZcndCfg.TrieCache, "Memory limit (MB) at which to flush the current in-memory trie to disk")
	falgs.DurationVar(&zconfig.ZcndCfg.TrieTimeout, "zcnd_trietimeout", zconfig.ZcndCfg.TrieTimeout, "Time limit after which to flush the current in-memory trie to disk")

	// txpool
	falgs.BoolVar(&zconfig.ZcndCfg.TxPool.NoLocals, "txpool_nolocals", zconfig.ZcndCfg.TxPool.NoLocals, "Disables price exemptions for locally submitted transactions")
	falgs.StringVar(&zconfig.ZcndCfg.TxPool.Journal, "txpool_journal", zconfig.ZcndCfg.TxPool.Journal, "Disk journal for local transaction to survive node restarts")
	falgs.DurationVar(&zconfig.ZcndCfg.TxPool.Rejournal, "txpool_rejournal", zconfig.ZcndCfg.TxPool.Rejournal, "Time interval to regenerate the local transaction journal")
	falgs.Uint64Var(&zconfig.ZcndCfg.TxPool.PriceBump, "txpool_pricebump", zconfig.ZcndCfg.TxPool.PriceBump, "Price bump percentage to replace an already existing transaction")
	falgs.Uint64Var(&zconfig.ZcndCfg.TxPool.PriceLimit, "txpool_pricelimit", zconfig.ZcndCfg.TxPool.PriceLimit, "Minimum gas price limit to enforce for acceptance into the pool")
	falgs.Uint64Var(&zconfig.ZcndCfg.TxPool.AccountSlots, "txpool_accountslots", zconfig.ZcndCfg.TxPool.AccountSlots, "Minimum number of executable transaction slots guaranteed per account")
	falgs.Uint64Var(&zconfig.ZcndCfg.TxPool.AccountQueue, "txpool_accountqueue", zconfig.ZcndCfg.TxPool.AccountQueue, "Maximum number of non-executable transaction slots permitted per account")
	falgs.Uint64Var(&zconfig.ZcndCfg.TxPool.GlobalSlots, "txpool_globalslots", zconfig.ZcndCfg.TxPool.GlobalSlots, "Maximum number of executable transaction slots for all accounts")
	falgs.Uint64Var(&zconfig.ZcndCfg.TxPool.GlobalQueue, "txpool_globalqueue", zconfig.ZcndCfg.TxPool.GlobalQueue, "Minimum number of non-executable transaction slots for all accounts")
	falgs.DurationVar(&zconfig.ZcndCfg.TxPool.Lifetime, "txpool_lifetime", zconfig.ZcndCfg.TxPool.Lifetime, "Maximum amount of time non-executable transaction are queued")

}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-1)
	}
}
