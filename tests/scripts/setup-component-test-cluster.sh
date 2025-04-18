#!/bin/bash

# This script is used to setup a system test cluster on a single machine.

# Function to print message and exit.
function error_exit {
  kubectl delete namespace monitoring 2> /dev/null
  kubectl delete namespace seclogic 2> /dev/null
  echo "$1" 1>&2
  exit 1
}

# Check that kubectl is installed.
if ! [ -x "$(command -v kubectl)" ]; then
  echo "kubectl is not installed. Please install kubectl and try again."
  exit 1
fi

# Check that either miniKube or kind is installed.
if ! [ -x "$(command -v minikube)" ] && ! [ -x "$(command -v kind)" ]; then
  echo "Either minikube or kind is not installed. Please install one of them and try again."
  exit 1
fi

# Check if the cluster is kind cluster by checking current context.
if [ "$(kubectl config current-context)" == "kind-kind" ]; then
  echo "Kind cluster detected."
  # Load the docker image into the kind cluster.
  kind load docker-image node-agent:latest || error_exit "Failed to load docker image into kind cluster."
fi

# Check if the cluster is minikube cluster by checking current context.
if [ "$(kubectl config current-context)" == "minikube" ]; then
  echo "Minikube cluster detected."
  # Load the docker image into the minikube cluster.
  minikube image load node-agent:latest || error_exit "Failed to load docker image into minikube cluster."
fi


# Check that helm is installed.
if ! [ -x "$(command -v helm)" ]; then
  echo "helm is not installed. Please install helm and try again."
  exit 1
fi

# Add prometheus helm repo and install prometheus.
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts || error_exit "Failed to add prometheus helm repo."
helm repo update || error_exit "Failed to update helm repos."
helm upgrade --install prometheus prometheus-community/kube-prometheus-stack \
    --namespace monitoring --create-namespace --wait --timeout 5m \
    --set grafana.enabled=true \
    --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false,prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false || error_exit "Failed to install prometheus."

# Check that the prometheus pod is running
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=prometheus -n monitoring --timeout=300s || error_exit "Prometheus did not start."

STORAGE_TAG=$(./storage-tag.sh)
NODE_AGENT_TAG=$(./node-agent-tag.sh)

# Install node agent chart
helm upgrade --install seclogic ../chart --set clusterName=`kubectl config current-context` \
    --set nodeAgent.image.tag=$NODE_AGENT_TAG \
    --set storage.image.tag=$STORAGE_TAG \
    -n seclogic --create-namespace --wait --timeout 5m || error_exit "Failed to install node-agent chart."

# Check that the node-agent pod is running
kubectl wait --for=condition=Ready pod -l app.kubernetes.io/name=node-agent -n seclogic --timeout=300s || error_exit "Node Agent did not start."

echo "System test cluster setup complete."

# port forward prometheus
# kubectl port-forward -n monitoring svc/prometheus-kube-prometheus-prometheus 9090:9090 &
# kubectl port-forward svc/alertmanager-operated 9093:9093 -n monitoring &
