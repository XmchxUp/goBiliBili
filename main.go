package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/xmchxup/goBiliBili/collect"
	"github.com/xmchxup/goBiliBili/logger"
	"github.com/xmchxup/goBiliBili/payload"
)

var uid string
var cookie string

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

	err := godotenv.Load()
	if err != nil {
		log.Error(ctx, "startup", "msg", err)
		os.Exit(1)
	}

	rootCmd.Flags().StringVar(&uid, "uid", "243824574", "BiliBili's uid")
	rootCmd.Flags().StringVar(&cookie, "cookie", os.Getenv("bilibili_cookie"), "BiliBili's cookie")

	rootCmd.AddCommand(versionCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Error(ctx, "startup", "msg", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger) error {

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	log.Info(ctx, "startup", "uid", uid)

	var f collect.Fetcher = &collect.BrowserFetch{
		Timeout: time.Duration(5000) * time.Millisecond,
		Log:     log,
	}

	// get all dynamic
	dynamicURLs, err := getUPDynamicURLByUID(uid, f)
	if err == nil {
		log.Info(ctx, "business", "dynamic_urls", dynamicURLs, "length", len(dynamicURLs))
	} else {
		log.Error(ctx, "business", "listUPDynamicURLByUID", err)
	}

	return nil
}

func getUPDynamicURLByUID(uid string, fetcher collect.Fetcher) ([]string, error) {
	baseURL := "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?host_mid=%s&timezone_offset=-480"

	var bilibiliResp payload.BiliBiliDynamicSimplifyResponse

	res := make([]string, 0, 10)

	for {
		url := fmt.Sprintf(baseURL, uid)
		if bilibiliResp.Data.Offset != "" {
			url += "&offset=" + bilibiliResp.Data.Offset
		}

		res = append(res, url)

		body, err := fetcher.Get(&collect.Request{URL: url, Cookie: cookie})
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(body, &bilibiliResp)
		if err != nil {
			return nil, err
		}

		if !bilibiliResp.Data.HasMore || bilibiliResp.Data.Offset == "" {
			break
		}

		time.Sleep(time.Duration(rand.Int31n(1000)+100) * time.Millisecond)
	}

	return res, nil
}
