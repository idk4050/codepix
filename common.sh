#!/bin/sh

export REGISTRY_NAME="codepix_registry"
export REGISTRY_PORT=2000

log() {
  echo "$(tput setaf 2)$(tput bold)$prefix: $1$(tput sgr 0)"
}
logerr() {
  echo "$(tput setaf 1)$(tput bold)$prefix: $1$(tput sgr 0)"
}

get_service_dirs() {
  find . -type f -name docker-compose.yml \
    | xargs -r dirname \
    | sed -e 's|^\./||'
}

read_env_file() {
  service_dir="$1"
  env_variables="$(2>/dev/null cat ./$service_dir/.env | grep ^[A-Za-z] | xargs -r)"
  set -a
  source <(printf "$env_variables")
  set +a
}

unset_env_file() {
  service_dir="$1"
  env_variables="$(2>/dev/null cat ./$service_dir/.env | grep ^[A-Za-z] | cut -d'=' -f1 | xargs -r)"
  unset "$env_variables"
}
