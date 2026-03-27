package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/osamikoyo/math-angel/internal/cash"
	"github.com/osamikoyo/math-angel/internal/config"
	"github.com/osamikoyo/math-angel/internal/handler"
	"github.com/osamikoyo/math-angel/internal/importer"
	"github.com/osamikoyo/math-angel/internal/repository"
	"github.com/osamikoyo/math-angel/internal/service"
	"github.com/osamikoyo/math-angel/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type App struct {
	echo     *echo.Echo
	httpSrv  *http.Server
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
	if err != nil {
		return nil, err
	}

	cache, err := setupCache(logger, cfg)
	if err != nil {
		return nil, err
	}

	service := service.NewService(repo, cache, cfg.Timeout)

	var importer *importer.Importer
	if cfg.Importer.Enabled {
		importer, err = setupImporter(service, logger, cfg)
		if err != nil {
			return nil, err
		}
	}

	e := setupEcho(service, logger)

	httpSrv := &http.Server{
		Addr:    cfg.Addr,
		Handler: e,
	}

	return &App{
		echo:    e,
		httpSrv: httpSrv,
		cfg:     cfg,
		importer: importer,
		logger:  logger,
	}, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGINT,
	)
	defer stop()

	if a.cfg.Importer.Enabled && a.importer != nil {
		a.importer.Start(ctx)
	}

	go func() {
		a.logger.Info("starting server", zap.String("addr", a.cfg.Addr))

		if err := a.httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Error("server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	a.logger.Info("received shutdown signal, gracefully stopping...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.shutdown(shutdownCtx); err != nil {
		a.logger.Error("graceful shutdown failed", zap.Error(err))
		return err
	}

	a.logger.Info("server stopped gracefully")
	return nil
}

func (a *App) shutdown(ctx context.Context) error {
	if err := a.httpSrv.Shutdown(ctx); err != nil {
		a.logger.Error("http server shutdown error", zap.Error(err))
	}

	return nil
}

func setupCfgAndLogger(configPath string) (*config.Config, *logger.Logger, error) {
	logger.Init(logger.Config{
		AppName:   "math-angel",
		AddCaller: false,
		LogFile:   "logs/math-angel.log",
		LogLevel:  "debug",
	})
	l := logger.Get()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		l.Error("failed load config", zap.String("path", configPath), zap.Error(err))
		return nil, nil, fmt.Errorf("failed load config: %w", err)
	}
	return cfg, l, nil
}

func setupRepo(logger *logger.Logger, cfg *config.Config) (*repository.Repository, error) {
	db, err := gorm.Open(sqlite.Open(cfg.DBpath))
	if err != nil {
		logger.Error("failed connect to db", zap.String("path", cfg.DBpath), zap.Error(err))
		return nil, fmt.Errorf("failed connect to db: %w", err)
	}
	logger.Info("database connected successfully")
	return repository.NewRepository(db, logger), nil
}

func setupCache(logger *logger.Logger, cfg *config.Config) (*cash.Cash, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		logger.Error("failed connect to redis", zap.String("addr", cfg.Redis.Addr), zap.Error(err))
		return nil, fmt.Errorf("failed connect to cache: %w", err)
	}

	logger.Info("redis connected successfully")
	return cash.NewCash(client, logger, cfg.Redis.ExpTime), nil
}

func setupImporter(service *service.Service, logger *logger.Logger, cfg *config.Config) (*importer.Importer, error) {
	importer, err := importer.NewImporter(service, cfg, logger)
	if err != nil {
		logger.Error("failed to setup importer", zap.Error(err))
		return nil, fmt.Errorf("failed setup importer: %w", err)
	}
	logger.Info("importer setup successfully")
	return importer, nil
}

func setupEcho(service *service.Service, logger *logger.Logger) *echo.Echo {
	e := echo.New()
	handler := handler.NewHandler(service)
	handler.RegisterRouters(e)

	logger.Info("echo configured successfully")
	return e
}