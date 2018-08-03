package main

import (
	"github.com/zipper-project/z0/node"
	"github.com/zipper-project/z0/zcnd"
)

type z0Config struct {
	ConfigFileFlag string
	NodeCfg        *node.Config
	ZcndCfg        *zcnd.Config
}
