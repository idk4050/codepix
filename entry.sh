#!/bin/sh

prefix="entry.sh"

cd $(dirname "$0")
source ./common.sh

entry() {
  log "Initializing swarm"
  docker swarm init 2>&1 | grep -v 'already part of a swarm'

  if [ -z "$(docker container ps | grep $REGISTRY_NAME)" ]; then
    log "Creating registry"
    docker run -d -p "$REGISTRY_PORT":5000 --restart=always --name "$REGISTRY_NAME" registry:2
  fi

  for service_dir in $(get_service_dirs); do
    stack_down="true"

    if [ -n "$(docker stack ps $service_dir -q 2> /dev/null)" ]; then
      stack_down=""
      log "$service_dir stack is up, attempting to remove it"
    fi

    until [ -n "$stack_down" ]; do
      if [ -n "$(docker stack ps $service_dir -q 2> /dev/null)" ]; then
        docker stack rm "$service_dir" 2>&1 | grep -v 'Failed to remove network'
        sleep 1
      else
        stack_down="true"
      fi
    done

    compose_file="$service_dir/docker-compose.yml"

    log "Reading $service_dir env file"
    read_env_file "$service_dir"

    log "Building $service_dir"
    docker-compose -f "$compose_file" build

    log "Pushing $service_dir to registry"
    docker-compose -f "$compose_file" push

    log "Deploying $service_dir"
    docker stack deploy --compose-file "$compose_file" "$service_dir"

    unset_env_file "$service_dir"
  done
}

entry
