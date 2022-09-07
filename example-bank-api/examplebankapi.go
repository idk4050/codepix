package examplebankapi

import (
	"codepix/example-bank-api/adapters/databaseclient"
	"codepix/example-bank-api/adapters/httputils"
	"codepix/example-bank-api/adapters/messagequeue"
	"codepix/example-bank-api/adapters/validator"
	"codepix/example-bank-api/config"
	customerdatabase "codepix/example-bank-api/customer/repository/database"
	customerservice "codepix/example-bank-api/customer/service"
	signupqueue "codepix/example-bank-api/customer/signup/queue"
	signuprepository "codepix/example-bank-api/customer/signup/repository"
	signupdatabase "codepix/example-bank-api/customer/signup/repository/database"
	signupservice "codepix/example-bank-api/customer/signup/service"
	userdatabase "codepix/example-bank-api/user/repository/database"
	signinqueue "codepix/example-bank-api/user/signin/queue"
	signinrepository "codepix/example-bank-api/user/signin/repository"
	signindatabase "codepix/example-bank-api/user/signin/repository/database"
	signinservice "codepix/example-bank-api/user/signin/service"
	"context"
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"path/filepath"

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
	logger           logr.Logger
	config           config.Config
	database         *databaseclient.Database
	messageQueue     *messagequeue.MessageQueue
	server           *http.Server
	signInRepository signinrepository.Repository
	signUpRepository signuprepository.Repository
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
	validator, err := validator.New()
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
	server := &http.Server{Addr: ":" + config.HTTP.Port, Handler: router}

	handle := httputils.RouterHandler(func(method string, path string, handler http.Handler) {
		router.Handler(method, filepath.Join("/api", path), handler)
	})
	apiDocsFS, err := fs.Sub(_apiDocsFS, "api-docs")
	if err != nil {
		return nil, err
	}
	apiDocsHandler := http.FileServer(http.FS(apiDocsFS))
	router.Handler("GET", "/api-docs/*filepath", chain.Then(
		http.StripPrefix("/api-docs", apiDocsHandler),
	))

	userRepository := userdatabase.Database{Database: database}
	signInRepository := signindatabase.Database{Database: database}
	customerRepository := customerdatabase.Database{Database: database}
	signUpRepository := signupdatabase.Database{Database: database}

	err = signinservice.Register(config, chain, handle, validator,
		messageQueue, signInRepository, userRepository, customerRepository)
	if err != nil {
		return nil, err
	}
	err = signupservice.Register(chain, handle, validator, messageQueue, signUpRepository)
	if err != nil {
		return nil, err
	}
	err = customerservice.Register(config, chain, handle, validator, customerRepository)
	if err != nil {
		return nil, err
	}

	api := &ExampleBankAPI{
		logger:           logger,
		config:           config,
		database:         database,
		messageQueue:     messageQueue,
		server:           server,
		signInRepository: signInRepository,
		signUpRepository: signUpRepository,
	}
	return api, nil
}

func (api ExampleBankAPI) Start(ctx context.Context) error {
	api.logger.Info("starting Example Bank API")

	err := api.database.AutoMigrate(
		&userdatabase.User{},
		&signindatabase.SignIn{},
		&customerdatabase.Customer{},
		&signupdatabase.SignUp{},
	)
	if err != nil {
		return err
	}
	err = signinqueue.SetupReaders(ctx, api.messageQueue, api.signInRepository)
	if err != nil {
		return err
	}
	err = signupqueue.SetupReaders(ctx, api.messageQueue, api.signUpRepository)
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
