package examplebankapi

import (
	"codepix/example-bank-api/adapters/databaseclient"
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/adapters/messagequeue"
	"codepix/example-bank-api/config"
	"context"
	"embed"
	"errors"
	"io/fs"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:embed api-docs
var _apiDocsFS embed.FS

type ExampleBankAPI struct {
	logger       logr.Logger
	config       config.Config
	database     *databaseclient.Database
	messageQueue *messagequeue.MessageQueue
	server       *http.Server
}

func New(ctx context.Context, loggerImpl *zap.Logger, config config.Config) (*ExampleBankAPI, error) {
	logger := zapr.NewLogger(loggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DPanicLevel),
		zap.WithCaller(false),
	))
	panicLogger := zapr.NewLogger(loggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DebugLevel),
		zap.AddCallerSkip(3),
		zap.Fields(
			zap.StackSkip("stacktrace", 3),
		),
	))
	database, err := databaseclient.Open(config, logger)
	if err != nil {
		return nil, err
	}
	messageQueue, err := messagequeue.Open(config, logger)
	if err != nil {
		return nil, err
	}

	router := httprouter.New()
	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = true
	router.HandleMethodNotAllowed = true

	chain := alice.New(
		httputils.PanicHandler(panicLogger),
		httputils.Logger(logger),
	)
	router.NotFound = chain.ThenFunc(httputils.NotFound)
	router.MethodNotAllowed = chain.ThenFunc(httputils.NotAllowed)

	apiDocsFS, err := fs.Sub(_apiDocsFS, "api-docs")
	if err != nil {
		return nil, err
	}
	apiDocsHandler := http.FileServer(http.FS(apiDocsFS))
	router.Handler("GET", "/api-docs/*filepath", chain.Then(
		http.StripPrefix("/api-docs", apiDocsHandler),
	))

	server := &http.Server{Addr: ":" + config.HTTP.Port, Handler: router}

	api := &ExampleBankAPI{
		logger:       logger,
		config:       config,
		database:     database,
		messageQueue: messageQueue,
		server:       server,
	}
	return api, nil
}

func (api ExampleBankAPI) Start(ctx context.Context) error {
	api.logger.Info("starting Example Bank API")

	err := api.database.AutoMigrate()
	if err != nil {
		return err
	}

	httpLogger := api.logger.WithName("http")
	httpLogger.Info("http server listening on port " + api.config.HTTP.Port)
	go func() {
		err = api.server.ListenAndServe()
		switch {
		case err == nil:
			return
		case errors.Is(err, http.ErrServerClosed):
			httpLogger.Info("http server closed")
		default:
			httpLogger.Error(err, "http server failed to serve")
		}
	}()

	api.logger.Info("Example Bank API started")
	return nil
}

func (api ExampleBankAPI) Stop() error {
	api.logger.Info("stopping Example Bank API")

	err := api.server.Shutdown(context.Background())
	if err != nil {
		api.logger.WithName("http").Error(err, "http server failed to shut down")
		return err
	}
	api.logger.WithName("http").Info("http server shut down")

	err = api.database.Close()
	if err != nil {
		return err
	}
	err = api.messageQueue.Close()
	if err != nil {
		return err
	}
	api.logger.Info("Example Bank API stopped")
	return nil
}
