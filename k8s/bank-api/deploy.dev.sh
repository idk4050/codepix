#!/bin/sh

cd $(dirname "$0")

namespace="codepix-bank-api"
kubectl create namespace $namespace

env_dir="../../bank-api/config/env"

for config_env_file in $env_dir/*.config.production.env.example; do
  config_map="$(basename "$config_env_file" | cut -d. -f1)"

  kubectl create configmap "$config_map" -n $namespace --from-env-file "$config_env_file" \
  || kubectl create configmap "$config_map" -n $namespace --from-env-file "$config_env_file" \
    -o yaml --dry-run=client | kubectl replace -f -
done
for secrets_env_file in $env_dir/*.secrets.production.env.example; do
  secret="$(basename "$secrets_env_file" | cut -d. -f1)"

  kubectl create secret generic "$secret" -n $namespace --from-env-file "$secrets_env_file" \
  || kubectl create secret generic "$secret" -n $namespace --from-env-file "$secrets_env_file" \
    -o yaml --dry-run=client | kubectl replace -f -
done

kubectl apply \
  -f api.yaml \
  -f database.yaml \
  -f eventstore.yaml \
  -f storeprojection.yaml \
  -f eventbus.yaml
