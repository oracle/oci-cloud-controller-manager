#!/bin/bash
#
# Move API machinery deps out of kubernetes.

rm -rf ./vendor/k8s.io/{apiserver,apimachinery,client-go}
cp -r ./vendor/k8s.io/kubernetes/staging/src/k8s.io/{apiserver,apimachinery,client-go} ./vendor/k8s.io
