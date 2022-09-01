package customerapi

import (
	"codepix/customer-api/adapters/databaseclient"
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/adapters/outbox"
	"codepix/customer-api/adapters/validator"
	"codepix/customer-api/config"
	"codepix/customer-api/lib/outboxes"
	"codepix/customer-api/lib/publishers"
	"codepix/customer-api/lib/repositories"
	"codepix/customer-api/user"
	userrepository "codepix/customer-api/user/repository"
	userdatabase "codepix/customer-api/user/repository/database"
	signineventhandler "codepix/customer-api/user/signin/eventhandler"
	signinemailsender "codepix/customer-api/user/signin/eventhandler/emailsender"
	signincommandhandler "codepix/customer-api/user/signin/interactor/commandhandler"
	signinpublisher "codepix/customer-api/user/signin/publisher"
	signindatabase "codepix/customer-api/user/signin/repository/database"
	signinservice "codepix/customer-api/user/signin/service"
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

type CustomerAPI struct {
	logger         logr.Logger
	config         config.Config
	database       *databaseclient.Database
	outbox         outboxes.Outbox
	server         *http.Server
	userRepository userrepository.Repository
}

func New(ctx context.Context, loggerImpl *zap.Logger, config config.Config) (*CustomerAPI, error) {
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
	publishers := map[outboxes.Namespace]publishers.Publisher{
		signineventhandler.Namespace: signinpublisher.Publisher{
			EventHandler: signinemailsender.EmailSender{},
		},
	}
	outbox, err := outbox.Open(config, logger, publishers)
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

	handle := httputils.RouterHandler(func(method string, path string, handler http.Handler) {
		router.Handler(method, filepath.Join("/api", path), handler)
	})

	chain := alice.New(
		httputils.PanicHandler(panicLogger),
		httputils.Logger(logger),
	)

	router.NotFound = chain.ThenFunc(httputils.NotFound)
	router.MethodNotAllowed = chain.ThenFunc(httputils.NotAllowed)

	userRepository := userdatabase.Database{Database: database}

	signInRepository := signindatabase.Database{Database: database, Outbox: outbox}
	signInInteractor := signincommandhandler.CommandHandler{
		Repository:     signInRepository,
		UserRepository: userRepository,
	}
	err = signinservice.Register(config, chain, handle, validator,
		signInInteractor, signInRepository, userRepository)
	if err != nil {
		return nil, err
	}

	apiDocsFS, err := fs.Sub(_apiDocsFS, "api-docs")
	if err != nil {
		return nil, err
	}
	apiDocsHandler := http.FileServer(http.FS(apiDocsFS))
	router.Handler("GET", "/api-docs/*filepath", chain.Then(
		http.StripPrefix("/api-docs", apiDocsHandler),
	))

	server := &http.Server{Addr: ":" + config.HTTP.Port, Handler: router}

	customerAPI := &CustomerAPI{
		config:         config,
		logger:         logger,
		database:       database,
		outbox:         outbox,
		server:         server,
		userRepository: userRepository,
	}
	return customerAPI, nil
}

func (api CustomerAPI) Start(ctx context.Context) error {
	api.logger.Info("starting customer API")

	err := api.database.AutoMigrate(
		&userdatabase.User{},
		&signindatabase.SignIn{},
	)
	if err != nil {
		return err
	}
	err = createInitialUsers(api.config, api.userRepository)
	if err != nil {
		return err
	}
	err = api.outbox.AutoMigrate()
	if err != nil {
		return err
	}
	go api.outbox.Start(ctx)

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

	api.logger.Info("customer API started")
	return nil
}

func (api CustomerAPI) Stop() error {
	api.logger.Info("stopping customer API")

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
	api.logger.Info("customer API stopped")
	return nil
}

func createInitialUsers(config config.Config, repository userrepository.Repository) error {
	emails := config.InitialState.UserEmails

	for _, email := range emails {
		err := repository.Exists(email)

		var notFound *repositories.NotFoundError
		if errors.As(err, &notFound) {
			_, err := repository.Add(user.User{
				Email: email,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
