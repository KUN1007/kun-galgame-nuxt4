package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/repository"
	"kun-galgame-api/internal/galgame/service"
	"kun-galgame-api/internal/infrastructure/cache"
	"kun-galgame-api/internal/infrastructure/database"
	"kun-galgame-api/pkg/config"
	"kun-galgame-api/pkg/logger"

	"github.com/joho/godotenv"
)

// Manual mirror of the in-app cron job. The forum runs the merge sync every 30
// minutes on its own; this exists for the one operator action the schedule has
// no answer for — -replay drops the cursor so the next drain walks catalog's
// entire merge history from the beginning and folds everything still foldable.
func main() {
	replay := flag.Bool("replay", false, "丢弃游标, 从 catalog 合并历史的开头重新拉取")
	flag.Parse()

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		slog.Error("加载配置失败", "error", err)
		os.Exit(1)
	}
	logger.Init(cfg.Server.Mode)

	db := database.NewPostgres(cfg.Database, cfg.Server.Mode)
	rdb := cache.NewRedis(cfg.Redis)
	gc := client.New(cfg.NextMoeAPI.BaseURL, cfg.NextMoeAPI.APIKey, cfg.NextMoeAPI.ImageCDNBase)

	if *replay {
		if err := rdb.Del(context.Background(), service.MergeCursorKey).Err(); err != nil {
			slog.Error("清除合并游标失败", "error", err)
			os.Exit(1)
		}
		slog.Info("已清除合并游标, 本次将重放全部合并历史")
	}

	service.NewGalgameMergeSync(gc, repository.NewGalgameMergeRepository(db), rdb).Run()
}
