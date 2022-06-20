#!/bin/sh

registry_hostname="registry.codepix.local"
registry_subnet="10.45.0.0/16"
registry_ip="10.45.0.2"
registry_port="80"

codepix_dir="$HOME/.local/share/containers/codepix"

# network
sudo podman stop registry-codepix
sudo podman rm registry-codepix

sudo podman network rm codepix
sudo podman network create --subnet="$registry_subnet" codepix

# registry
registry_dir="$codepix_dir/registry"
mkdir -p "$registry_dir"

sudo podman run --detach --restart always \
  -v "$registry_dir":/var/lib/registry \
  -e "REGISTRY_HTTP_ADDR=0.0.0.0:$registry_port" \
  --network codepix --hostname "$registry_hostname" --ip "$registry_ip" -p "$registry_port:5000" \
  --name registry-codepix \
  docker.io/registry:2

if [ -z "$(grep $registry_ip /etc/hosts)" ]; then
  echo "$registry_ip  $registry_hostname" | sudo tee -a /etc/hosts
fi

# k8s setup
namespace="kube-system"

kubectl delete secret regcred -n $namespace
kubectl create secret generic regcred -n $namespace \
  --from-file=.dockerconfigjson="$auth_file" \
  --type=kubernetes.io/dockerconfigjson

# k3s setup
if [ -d /etc/rancher/k3s ]; then
  echo "
mirrors:
  $registry_hostname:
    endpoint:
      - \"http://$registry_hostname:$registry_port\"
" | sudo tee /etc/rancher/k3s/registries.yaml
fi
