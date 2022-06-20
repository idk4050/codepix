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
AUTH_HEADER='auth-token: eyJhbGciOiJSUzUxMiIsInR5cCI6IkpXVCJ9.eyJiYW5rX2lkIjoiMzg2ZjhmMjctYjI4OC00YjNiLWI2ZDEtOTMyNTAyMDAzNzZhIiwiZXhwIjoxNzc4NzgyNjU5LCJpYXQiOjE2NTg3ODI2NTksIm5iZiI6MTY1ODc4MjY1OX0.r-5U7LF2KGok8OAJxqNmv4_vxkr98wt0-JtEWhbrNokQO971eesgVTuTr5kwFbsBOC5Y1ixrRO2SLIlSjwAHbnpQCWDBNIlWzBXOVNHYQJOo1ziZjmnxvJf_Br0U9ZCrbODnm1BxN4nLfVtopgVucnkoGhtadwD5WM7bD4Svo24'

grpcui -plaintext -reflect-header "$AUTH_HEADER" -rpc-header "$AUTH_HEADER" localhost:4000
```
