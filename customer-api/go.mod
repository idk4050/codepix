module codepix/customer-api

go 1.19

require (
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/go-logr/logr v0.4.0
	github.com/go-logr/zapr v0.4.0
	github.com/go-playground/locales v0.14.0
	github.com/go-playground/mold/v4 v4.2.0
	github.com/go-playground/universal-translator v0.18.0
	github.com/go-playground/validator/v10 v10.11.0
	github.com/golang-jwt/jwt v3.2.2+incompatible
	github.com/google/uuid v1.3.0
	github.com/iancoleman/strcase v0.2.0
	github.com/jackc/pgconn v1.13.0
	github.com/jackc/pgerrcode v0.0.0-20220416144525-469b46aa5efa
	github.com/jonboulle/clockwork v0.3.0
	github.com/julienschmidt/httprouter v1.3.0
	github.com/justinas/alice v1.2.0
	github.com/mattn/go-sqlite3 v1.14.15
	github.com/mcuadros/go-lookup v0.0.0-20200831155250-80f87a4fa5ee
	github.com/mitchellh/mapstructure v1.5.0
	github.com/omaskery/outboxen v0.4.0
	github.com/omaskery/outboxen-gorm v0.4.0
	github.com/stretchr/testify v1.8.0
	github.com/subosito/gotenv v1.4.1
	go.uber.org/zap v1.23.0
	golang.org/x/text v0.3.7
	gorm.io/driver/postgres v1.3.9
	gorm.io/driver/sqlite v1.3.6
	gorm.io/gorm v1.23.8
)

replace github.com/omaskery/outboxen-gorm => github.com/idk4050/outboxen-gorm v0.4.1

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.12.0 // indirect
	github.com/jackc/pgx/v4 v4.17.1 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/segmentio/go-camelcase v0.0.0-20160726192923-7085f1e3c734 // indirect
	github.com/segmentio/go-snakecase v1.2.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90 // indirect
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
