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
