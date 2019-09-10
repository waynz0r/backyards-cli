#!/bin/bash
CHARTS_DIR=${1:-assets/charts}
ASSETS_DIR="${2:-assets}"

BACKYARDS_CHART_VERSION="0.1.11"
ISTIO_OPERATOR_CHART_VERSION="0.0.16"
CANARY_OPERATOR_CHART_VERSION="0.1.3"
BACKYARDS_DEMO_CHART_VERSION="0.1.0"
CERT_MANAGER_CHART_VERSION="v0.10.0"
CERT_MANAGER_CRDS="https://raw.githubusercontent.com/jetstack/cert-manager/release-0.10/deploy/manifests/00-crds.yaml"

mkdir -p ${CHARTS_DIR};

CHARTS=("https://kubernetes-charts.banzaicloud.com/charts/istio-operator-${ISTIO_OPERATOR_CHART_VERSION}.tgz")
CHARTS+=("https://kubernetes-charts.banzaicloud.com/charts/canary-operator-${CANARY_OPERATOR_CHART_VERSION}.tgz")
CHARTS+=("https://kubernetes-charts.banzaicloud.com/charts/backyards-${BACKYARDS_CHART_VERSION}.tgz")
CHARTS+=("https://kubernetes-charts.banzaicloud.com/charts/backyards-demo-${BACKYARDS_DEMO_CHART_VERSION}.tgz")
CHARTS+=("https://kubernetes-charts.banzaicloud.com/charts/backyards-demo-${BACKYARDS_DEMO_CHART_VERSION}.tgz")
CHARTS+=("https://kubernetes-charts.banzaicloud.com/charts/backyards-demo-${BACKYARDS_DEMO_CHART_VERSION}.tgz")
CHARTS+=("https://charts.jetstack.io/charts/cert-manager-${CERT_MANAGER_CHART_VERSION}.tgz")

for i in ${CHARTS[@]}; do
    curl -s "${i}" | tar -zxv --directory "${CHARTS_DIR}/" -f -
    retVal=$?
    if [ $retVal -ne 0 ]; then
        exit $retVal
    fi
done

curl -s "${CERT_MANAGER_CRDS}" -o "${ASSETS_DIR}/cert-manager/crds.yaml"
retVal=$?
if [ $retVal -ne 0 ]; then
    exit $retVal
fi

find "${CHARTS_DIR}/" -exec touch -t 201901010101 {} +
find "${ASSETS_DIR}/" -exec touch -t 201901010101 {} +

