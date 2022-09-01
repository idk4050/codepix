# Customer API

## Development

To build and push the customer-api image, use [/k8s/customer-api/build.dev.sh](../k8s/customer-api/build.dev.sh)

To deploy the API, use [/k8s/customer-api/deploy.dev.sh](../k8s/customer-api/deploy.dev.sh)

To delete the deploy and its resources, use `kubectl delete namespaces codepix-customer-api`

To run tests, you'll need to create an env file. Run the [create-env-file.sh](./config/create-env-file.sh) script with `test` or `production` as the only argument. A set of `.env.example` files will be selected accordingly to create the resulting `.env` file.

Example command: 
```
./config/create-env-file.sh test
```

<br>

## Browsing the API

To browse the API, use [vscode-restclient](https://github.com/Huachao/vscode-restclient). The http sheet is inside the [api-docs/](api-docs/vscode-rest-client.http) directory, which is also statically served on the public endpoint `/api-docs/`

<br>

## Authentication

Authentication is done via JWT.

Sign-up and sign-in use single-use tokens, which are printed to the terminal rather than being sent via email for the time being.

### User authentication

Validation keys can be rotated through the `USER_AUTH_PREVIOUS_VALIDATION_KEY` variable. During authentication, it will be tried if the current one fails.

### Bank authentication

Banks can also use their API keys to request tokens for use in the Bank API.

Token signing is controlled by the Customer API while token validation is done by the Bank API.

See the Bank API [README](../bank-api/README.md#authentication) for more information.
