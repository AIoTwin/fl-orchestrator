#!/bin/bash

if [ -z "$1" ]; then
  echo "Usage: $0 <namespace>"
  exit 1
fi

NAMESPACE=$1

echo "Deleting all resources in namespace: $NAMESPACE"

kubectl -n "$NAMESPACE" delete deployments --all
kubectl -n "$NAMESPACE" delete configmap --all
kubectl -n "$NAMESPACE" delete svc --all