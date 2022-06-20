#!/bin/sh

prefix="clean.sh"

cd $(dirname "$0")
source ./common.sh

clean() {
  log "Stopping registry"
  docker container stop "$REGISTRY_NAME"
  log "Removing registry"
  docker container rm -v "$REGISTRY_NAME"

  for service_dir in $(get_service_dirs); do
    log "Removing $service_dir stack"
    docker stack rm "$service_dir"

    compose_file="$service_dir/docker-compose.yml"

    read_env_file "$service_dir"

    containers_still_up="$(docker-compose -f "$compose_file" ps -aq)"

    if [ -n "$containers_still_up" ]; then
      log "Waiting for $service_dir containers to be removed"

      until [ -z "$containers_still_up" ]; do
        containers_still_up="$(docker-compose -f "$compose_file" ps -aq)"
        sleep 1
      done
    fi

    log "Removing images from $service_dir"
    docker-compose -f "$compose_file" down --rmi all

    unset_env_file "$service_dir"
  done

  log "Leaving swarm"
  docker swarm leave --force

  log "Removing volumes from stack"
  docker volume ls -q | grep "${service_dir}_" | xargs -r docker volume rm
}

clean
