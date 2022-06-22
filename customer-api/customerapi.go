package customerapi

import (
	"codepix/customer-api/adapters/databaseclient"
	"codepix/customer-api/adapters/httputils"
	"codepix/customer-api/adapters/outbox"
	"codepix/customer-api/adapters/validator"
	"codepix/customer-api/config"
	apikeyusecase "codepix/customer-api/customer/bank/apikey/interactor/usecase"
	apikeydatabase "codepix/customer-api/customer/bank/apikey/repository/database"
	apikeyservice "codepix/customer-api/customer/bank/apikey/service"
	bankauth "codepix/customer-api/customer/bank/auth"
	bankusecase "codepix/customer-api/customer/bank/interactor/usecase"
	bankdatabase "codepix/customer-api/customer/bank/repository/database"
	bankservice "codepix/customer-api/customer/bank/service"
	customerdatabase "codepix/customer-api/customer/repository/database"
	customerservice "codepix/customer-api/customer/service"
	signupeventhandler "codepix/customer-api/customer/signup/eventhandler"
	signupemailsender "codepix/customer-api/customer/signup/eventhandler/emailsender"
	signupcommandhandler "codepix/customer-api/customer/signup/interactor/commandhandler"
	signuppublisher "codepix/customer-api/customer/signup/publisher"
	signupdatabase "codepix/customer-api/customer/signup/repository/database"
	signupservice "codepix/customer-api/customer/signup/service"
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
	"gorm.io/gorm"
)

//go:embed api-docs
var _apiDocsFS embed.FS

type CustomerAPI struct {
	config         config.Config
	logger         logr.Logger
	database       *gorm.DB
	outbox         outboxes.Outbox
	server         *http.Server
	userRepository userrepository.Repository
}

func New(config config.Config, loggerImpl *zap.Logger) (*CustomerAPI, error) {
	logger := zapr.NewLogger(loggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DPanicLevel),
		zap.WithCaller(false),
	))
	database, err := databaseclient.Open(config, logger)
	if err != nil {
		return nil, err
	}
	publishers := map[outboxes.Namespace]publishers.Publisher{
		signupeventhandler.Namespace: signuppublisher.Publisher{
			EventHandler: signupemailsender.EmailSender{},
		},
		signineventhandler.Namespace: signinpublisher.Publisher{
			EventHandler: signinemailsender.EmailSender{},
		},
	}
	outbox, err := outbox.New(config, logger, publishers)
	if err != nil {
		return nil, err
	}
	chain := alice.New(
		httputils.Logger(logger),
	)
	panicLogger := zapr.NewLogger(loggerImpl.WithOptions(
		zap.AddStacktrace(zapcore.DebugLevel),
		zap.AddCallerSkip(3),
		zap.Fields(
			zap.StackSkip("stacktrace", 3),
		),
	))
	router := httprouter.New()
	router.PanicHandler = httputils.PanicLogger(panicLogger)
	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = true

	apiDocsFS, err := fs.Sub(_apiDocsFS, "api-docs")
	if err != nil {
		return nil, err
	}
	router.ServeFiles("/api-docs/*filepath", http.FS(apiDocsFS))

	handle := httputils.RouterHandler(func(method string, path string, handler http.Handler) {
		router.Handler(method, filepath.Join("/api", path), handler)
	})

	validator, err := validator.New()
	if err != nil {
		return nil, err
	}

	signUpRepository := signupdatabase.Database{DB: database, Outbox: outbox}
	signUpInteractor := signupcommandhandler.CommandHandler{Repository: signUpRepository}
	_, err = signupservice.New(signUpInteractor, chain, handle, validator)
	if err != nil {
		return nil, err
	}

	userRepository := userdatabase.Database{DB: database}
	customerRepository := customerdatabase.Database{DB: database}

	signInRepository := signindatabase.Database{DB: database, Outbox: outbox}
	signInInteractor := signincommandhandler.CommandHandler{
		Repository:     signInRepository,
		UserRepository: userRepository,
	}
	_, err = signinservice.New(config, chain, handle, validator,
		signInInteractor, signInRepository, userRepository, customerRepository)
	if err != nil {
		return nil, err
	}

	bankRepository := bankdatabase.Database{DB: database}
	_, err = customerservice.New(config, chain, handle, validator, customerRepository, bankRepository)
	if err != nil {
		return nil, err
	}

	apiKeyRepository := apikeydatabase.Database{DB: database}
	bankInteractor := bankusecase.Usecase{
		Repository:         bankRepository,
		CustomerRepository: customerRepository,
	}
	_, err = bankservice.New(config, chain, handle, validator,
		bankInteractor, bankRepository, apiKeyRepository)
	if err != nil {
		return nil, err
	}

	apiKeyInteractor := &apikeyusecase.Usecase{Repository: apiKeyRepository}
	_, err = apikeyservice.New(config, chain, handle, validator,
		apiKeyInteractor, apiKeyRepository, bankRepository)
	if err != nil {
		return nil, err
	}

	err = bankauth.New(config, chain, handle, validator, apiKeyRepository)
	if err != nil {
		return nil, err
	}

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
	err := api.database.AutoMigrate(
		&signupdatabase.SignUp{},
		&signindatabase.SignIn{},
		&userdatabase.User{},
		&customerdatabase.Customer{},
		&bankdatabase.Bank{},
		&apikeydatabase.APIKey{},
	)
	if err != nil {
		return err
	}

	err = api.outbox.AutoMigrate()
	if err != nil {
		return err
	}
	go api.outbox.Start(ctx)

	err = seedInitialUsers(api.config, api.userRepository)
	if err != nil {
		return err
	}

	go func() {
		api.logger.Info("server listening on port " + api.config.HTTP.Port)

		err = api.server.ListenAndServe()
		switch {
		case err == nil:
			return
		case errors.Is(err, http.ErrServerClosed):
			api.logger.Info("server closed")
		default:
			api.logger.Error(err, "server failed to serve")
		}
	}()
	return nil
}

func (api CustomerAPI) Stop() error {
	err := api.server.Shutdown(context.Background())
	if err != nil {
		api.logger.Error(err, "server failed to shut down")
		return err
	}
	api.logger.Info("server shut down")
	return nil
}

func seedInitialUsers(config config.Config, repository userrepository.Repository) error {
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
