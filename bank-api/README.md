# Bank API

## Development

To create an env file from the example env files, run the `create-env-file.sh` script in `config/` with `test` or `production` as the only argument. A set of `.env.example` files will be selected accordingly to create the resulting `.env` file. The `.env` file will be loaded at runtime but it won't override existing variables.

Example command: 
```
./config/create-env-file.sh test
```

### Test environment

The `test` env files are meant to be used for quick testing, manual or automated. Databases are setup in memory and the program may behave differently than in `production`.

In the `test` environment, the event store adapter requires a mongodb binary to run 'in memory'. After creating the `.env` file, you'll need to change the `ES_IN_MEMORY_BINARY_PATH` variable manually. The same applies to the store projection with the variable `SP_IN_MEMORY_BINARY_PATH`.

### Production environment

The `production` env files are meant to be used for testing with Docker. See the main [README](../README.md) file for instructions on how to setup the Docker services.

To run tests in the `production` environment, you can use:
```
docker container exec -it <container_id> go test ./...
```

To debug in the `production` environment, you can edit the resulting `.env` as follows:

1. set the hostnames of other service dependencies from `localhost` to your local Docker address (usually `172.17.0.1`)
2. change the port of the service you are debugging to prevent conflict with the existing instance

Now you can debug the service locally with `production` dependencies.

<br>

## Browsing the API

To browse the API, use [grpc-ui](https://github.com/fullstorydev/grpcui). Example command:

```
grpcui -plaintext localhost:4000
```

<br>

## Authentication

Authentication is done via JWT.

The `BANK_AUTH_VALIDATION_KEY` must be derived from the bank auth signing key in the Customer API.

The validation keys can be rotated through the `BANK_AUTH_PREVIOUS_VALIDATION_KEY` variable. During authentication, it will be tried if the current one fails.

Banks use their API keys (created in the Customer API) to obtain tokens from the Customer API.

Tokens obtained from the Customer API contain all the information required by the bank for use in the Bank API. The databases are not shared.

See the Customer API [README](../customer-api/README.md#authentication) for more information.
