#!/bin/bash

# delete old resource
kubectl delete deployment webhook-controller-manager -n webhook-system
kubectl delete book book-sample

# make manifests
make manifests

# make install
make install

# make docker-build image
make docker-build

# save image into tar
docker save -o controller:latest.tar controller:latest

# load image into minikube
minikube image load controller:latest.tar

# make deploy
IMG=docker.io/library/controller:latest make deploy
