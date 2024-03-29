package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/user"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/XmchxUp/goBiliBili/collect"
	"github.com/XmchxUp/goBiliBili/logger"
	"github.com/XmchxUp/goBiliBili/payload"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var uid string
var cookie string
var baseDir string

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

	err := godotenv.Load()
	if err != nil {
		log.Warn(ctx, "startup", "msg", err)
	}

	usr, _ := user.Current()

	baseDir = fmt.Sprintf("%v/Pictures/goBiliBili", usr.HomeDir)
	err = createDirIfNotExist(baseDir)
	if err != nil {
		return err
	}

	log.Info(ctx, "startup", "download directory path", baseDir)

	var f collect.Fetcher = &collect.BrowserFetch{
		Timeout: time.Duration(10000) * time.Millisecond,
		Log:     log,
	}

	// get all dynamic urls
	workerCnt := 100
	picDataCh := make(chan *PicDownloadData, workerCnt)
	for i := 0; i < 100; i++ {
		go downloadWorker(picDataCh, f, log)
	}

	if err := parseUPDynamic(uid, f, picDataCh); err != nil {
		return err
	}

	return nil
}

func createDirIfNotExist(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return err
		}

	} else if err != nil {
		return err
	}

	return nil
}

func downloadWorker(ch chan *PicDownloadData, fetcher collect.Fetcher, log *logger.Logger) {
	var imageType string
	for d := range ch {
		for i := 0; i < len(d.URLs); i++ {
			if strings.Contains(d.URLs[i], ".png") {
				imageType = ".png"
			} else {
				imageType = ".jpg"
			}

			body, err := fetcher.Get(&collect.Request{URL: d.URLs[i]})
			if err != nil {
				log.Error(context.Background(), "download-worker", "fetcher err", err)
				continue
			}

			filePath := fmt.Sprintf("%s/%s%d%s", baseDir, d.Name, i, imageType)

			if _, err := os.Stat(filePath); err == nil {
				continue
			}

			err = os.WriteFile(filePath, body, 0644)
			if err != nil {
				log.Error(context.Background(), "download-worker", "write file err", err)
				continue
			}
			log.Info(context.Background(), "download", "status", "done", "info", d)
		}
	}
}

type PicDownloadData struct {
	Name string
	URLs []string
}

func removeWhitespace(str string) string {
	re := regexp.MustCompile(`\s+`)
	return re.ReplaceAllString(str, "")
}

func parseUPDynamic(uid string, fetcher collect.Fetcher, picDataCh chan *PicDownloadData) error {
	baseURL := "https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?host_mid=%s&timezone_offset=-480"

	var bilibiliResp payload.BiliBiliDynamicResponse

	now := time.Now()
	year, month, day := now.Date()

	for {
		url := fmt.Sprintf(baseURL, uid)
		if bilibiliResp.Data.Offset != "" {
			url += "&offset=" + bilibiliResp.Data.Offset
		}

		body, err := fetcher.Get(&collect.Request{URL: url, Cookie: cookie})
		if err != nil {
			return err
		}

		err = json.Unmarshal(body, &bilibiliResp)
		if err != nil {
			return err
		}

		for _, item := range bilibiliResp.Data.Items {
			name := item.Modules.ModuleAuthor.Name
			desc := item.Modules.ModuleDynamic.Major.Archive.Desc
			title := item.Modules.ModuleDynamic.Major.Archive.Title

			picURLs := make([]string, 0, len(item.Modules.ModuleDynamic.Major.Archive.Pics)+1)

			if item.Modules.ModuleDynamic.Major.Archive.Cover != "" {
				picURLs = append(picURLs, item.Modules.ModuleDynamic.Major.Archive.Cover)
			}

			for _, pic := range item.Modules.ModuleDynamic.Major.Archive.Pics {
				picURLs = append(picURLs, pic.URL)
			}

			fileName := fmt.Sprintf("%d-%02d-%02d%s-%s-%s", year, month, day, name, title, desc)

			picDataCh <- &PicDownloadData{
				Name: removeWhitespace(fileName),
				URLs: picURLs,
			}
		}

		if !bilibiliResp.Data.HasMore || bilibiliResp.Data.Offset == "" {
			break
		}

		time.Sleep(time.Duration(rand.Int31n(1000)+100) * time.Millisecond)
	}

	return nil
}
