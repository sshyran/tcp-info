package collector

import (
	"context"

	"github.com/m-lab/tcp-info/netlink"
	"github.com/m-lab/tcp-info/saver"
)

// No code, but needed for compiling.

// Run is an empty implementation to allow compiling code in darwin.
func Run(ctx context.Context, reps int, svrChan chan<- []*netlink.ArchivalRecord, cl saver.CacheLogger, skipLocal bool) (localCount, errCount int) {
	// Does notihg in Darwin
	return 0, 0
}