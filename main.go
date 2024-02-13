package main

import (
	"context"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/xmchxup/goBiliBili/logger"
)

var uid string

func main() {
	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "******* SEND ALERT ******")
		},
	}

	traceIDFn := func(ctx context.Context) string {
		return "ultraman-test-trace-id"
	}

	ctx := context.Background()
	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "goBiliBili", traceIDFn, events)

	var rootCmd = &cobra.Command{
		Use:   "goBiliBili",
		Short: "goBiliBili  is the tool used to download stuff related to the B site",
		Long:  `goBiliBili  is the tool used to download stuff related to the B site`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := run(ctx, log); err != nil {
				log.Error(ctx, "startup", "msg", err)
				os.Exit(1)
			}
		},
	}
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "v1.0",
		Long:  "v1.0",
	}

	rootCmd.Flags().StringVar(&uid, "uid", "243824574", "BiliBili's uid")

	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Error(ctx, "startup", "msg", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger) error {
	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	log.Info(ctx, "startup", "uid", uid)

	return nil
}
