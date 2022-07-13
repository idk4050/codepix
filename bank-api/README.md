# Bank API

## Development

To build and push the bank-api image, use [/k8s/bank-api/build.dev.sh](../k8s/bank-api/build.dev.sh)

To deploy the API, use [/k8s/bank-api/deploy.dev.sh](../k8s/bank-api/deploy.dev.sh)

To delete the deploy and its resources, use `kubectl delete namespaces codepix-bank-api`

To run tests, you'll need to create an env file. Run the [create-env-file.sh](./config/create-env-file.sh) script with `test` or `production` as the only argument. A set of `.env.example` files will be selected accordingly to create the resulting `.env` file.

Example command: 
```
./config/create-env-file.sh test
```

<br>

## Browsing the API

To browse the API, use [grpc-ui](https://github.com/fullstorydev/grpcui). Example command:

```
AUTH_HEADER='authorization: jjbgdLoZQMYoBtaQ5r8TTB4aXba86zAlYsELs0TDNMcKjy50bJc1iByZ7p6aTgQ0Pa1EZbdazWaubhwJ6LkTS3mMoAAmuMcnvXvR'

grpcui -plaintext -reflect-header "$AUTH_HEADER" -rpc-header "$AUTH_HEADER" localhost:4000
```

<br>

## Authentication

Authentication is done via JWT.

The `BANK_AUTH_VALIDATION_KEY` must be derived from the bank auth signing key in the Customer API.

The validation keys can be rotated through the `BANK_AUTH_PREVIOUS_VALIDATION_KEY` variable. During authentication, it will be tried if the current one fails.

Banks use their API keys (created in the Customer API) to obtain tokens from the Customer API.

Tokens obtained from the Customer API contain all the information required by the bank for use in the Bank API. The databases are not shared.

See the Customer API [README](../customer-api/README.md#authentication) for more information.
