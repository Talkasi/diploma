package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"logistics/internal/infrastructure/config"
	"logistics/internal/infrastructure/db"
	"logistics/internal/infrastructure/logger"
	"logistics/internal/infrastructure/osm"
	"logistics/internal/infrastructure/repository"
	"logistics/internal/usecase/importer"
)

func main() {
	if err := loadEnv(); err != nil {
		log.Printf("предупреждение: %v", err)
	}

	params, err := parseFlags()
	if err != nil {
		log.Fatalf("%v", err)
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}
	logr, err := logger.New(appEnv)
	if err != nil {
		log.Fatalf("ошибка инициализации логгера: %v", err)
	}
	defer logr.Sync()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	start := time.Now()

	osmReader := osm.NewPBFReader()

	// Dry-run не требует БД.
	var svc *importer.Service
	if params.DryRun {
		svc = importer.NewService(osmReader, nil, nil)
	} else {
		cfg, err := config.Load()
		if err != nil {
			log.Fatalf("ошибка загрузки конфига: %v", err)
		}
		pool, err := db.NewPool(ctx, cfg, logr)
		if err != nil {
			logr.Fatal("не удалось подключиться к БД", zap.Error(err))
		}
		defer pool.Close()
		repo := repository.NewImportRepository(pool)
		svc = importer.NewService(osmReader, repo, repo)
	}

	stats, err := svc.Import(ctx, params)
	if err != nil {
		logr.Fatal("импорт завершился ошибкой", zap.Error(err))
	}

	logr.Info("импорт завершён", zap.Duration("elapsed", time.Since(start)))
	logr.Info("статистика",
		zap.Int("ways_read", stats.Parse.WaysRead),
		zap.Int("ways_accepted", stats.Parse.WaysAccepted),
		zap.Int("nodes_needed", stats.Parse.NodesNeeded),
		zap.Int("nodes_found", stats.Parse.NodesFound),
		zap.Int("nodes_inserted", stats.Write.NodesInserted),
		zap.Int("edges_built", stats.Parse.EdgesBuilt),
		zap.Int("edges_inserted", stats.Write.EdgesInserted),
		zap.Int("skipped_missing_node", stats.Parse.SkippedEdgesMissingNode),
		zap.Int("skipped_out_bbox", stats.Parse.SkippedEdgesOutOfBBox),
		zap.Int("skipped_zero_length", stats.Parse.SkippedEdgesZeroLength),
		zap.Duration("parse1", stats.DurationParse1),
		zap.Duration("parse2", stats.DurationParse2),
		zap.Duration("build", stats.DurationBuild),
		zap.Duration("write", stats.DurationWrite),
	)
}

func loadEnv() error {
	primary := "../../infra/.env"
	if _, err := os.Stat(primary); err == nil {
		return godotenv.Load(primary)
	}
	return godotenv.Load()
}

func parseFlags() (importer.ImportParams, error) {
	var params importer.ImportParams
	var bboxStr string
	var bboxPad float64

	flag.StringVar(&params.PBFPath, "pbf", "", "путь к .osm.pbf файлу (обязательно)")
	flag.StringVar(&params.City, "city", "", "название города (обязательно)")
	flag.StringVar(&bboxStr, "bbox", "", "границы bbox в формате minLat,minLon,maxLat,maxLon (обязательно)")
	flag.Float64Var(&bboxPad, "bbox-pad", 0, "дополнительный паддинг к bbox (в градусах)")
	flag.StringVar(&params.Source, "source", "local-pbf", "источник данных")
	flag.BoolVar(&params.DryRun, "dry-run", false, "не записывать в БД, только считать статистику")
	flag.IntVar(&params.LimitWays, "limit-ways", 0, "ограничить количество way для отладки")
	flag.StringVar(&params.HighwayProfile, "highway-profile", "car", "профиль дорог (пока поддерживается только car)")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Использование: %s --pbf=FILE --city=NAME --bbox=minLat,minLon,maxLat,maxLon [опции]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	// Валидация.
	if strings.TrimSpace(params.PBFPath) == "" || strings.TrimSpace(params.City) == "" || strings.TrimSpace(bboxStr) == "" {
		flag.Usage()
		return params, fmt.Errorf("обязательны флаги --pbf, --city, --bbox")
	}

	bbox, err := importer.ParseBBox(bboxStr, bboxPad)
	if err != nil {
		return params, fmt.Errorf("bbox: %w", err)
	}
	params.BBox = bbox

	return params, nil
}
