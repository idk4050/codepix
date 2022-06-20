# Bank API

## Development

To create an env file from the example env files, run the `create-env-file.sh` script in `config/` with `test` or `production` as the only argument. A set of `.env.example` files will be selected accordingly to create the resulting `.env` file. The `.env` file will be loaded at runtime but it won't override existing variables.

Example command: 
```
./config/create-env-file.sh test
```

### Test environment

The `test` env files are meant to be used for quick testing, manual or automated. Databases are setup in memory and the program may behave differently than in `production`.

In the `test` environment, the event store adapter requires a mongodb binary to run 'in memory'. After creating the `.env` file, you'll need to change the `ES_IN_MEMORY_BINARY_PATH` variable manually.

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
