package app

import (
	"fmt"

	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/cash"
	"github.com/osamikoyo/math-angel/internal/config"
	"github.com/osamikoyo/math-angel/internal/importer"
	"github.com/osamikoyo/math-angel/internal/repository"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type App struct {
	echo     *echo.Echo
	cfg      *config.Config
	importer *importer.Importer
	logger   *logger.Logger
}

func SetupApp(configPath string) (*App, error) {
	cfg, logger, err := setupCfgAndLogger(configPath)
	if err != nil {
		return nil, err
	}

	repo, err := setupRepo(logger, cfg)
	if err != nil{
		return nil, err
	}
}

func setupCfgAndLogger(configPath string) (*config.Config, *logger.Logger, error) {
	logger.Init(logger.Config{
		AppName:   "math-angel",
		AddCaller: false,
		LogFile:   "logs/math-angel.log",
		LogLevel:  "debug",
	})

	logger := logger.Get()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Error("failed load config",
			zap.String("path", configPath),
			zap.Error(err))

		return nil, nil, fmt.Errorf("failed load config: %w", err)
	}

	return cfg, logger, nil
}

func setupRepo(logger *logger.Logger, cfg *config.Config) (*repository.Repository, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DBpath))
	if err != nil {
		logger.Error("failed connect to db",
			zap.String("path", cfg.DBpath),
			zap.Error(err))

		return nil, fmt.Errorf("failed connect to db: %w", err)
	}

	logger.Info("successfully connect to db")

	return repository.NewRepository(db, logger), nil
}

func setupCash(logger *logger.Logger, cfg *config.Config) (*cash.Cash, error) {
	client := redis.NewClient()
}