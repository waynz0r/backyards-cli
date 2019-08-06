#!/bin/bash
CHARTS_DIR=${1:-.charts}

BACKYARDS_CHART_VERSION="0.1.1"
ISTIO_OPERATOR_CHART_VERSION="0.0.16"
CANARY_OPERATOR_CHART_VERSION="0.1.2"
BACKYARDS_DEMO_CHART_VERSION="0.1.0"

mkdir -p ${CHARTS_DIR};
curl -s https://kubernetes-charts.banzaicloud.com/charts/istio-operator-${ISTIO_OPERATOR_CHART_VERSION}.tgz | tar -zxv --directory ${CHARTS_DIR}/ -f -
retVal=$?
if [ $retVal -ne 0 ]; then
    exit $retVal
fi

curl -s https://kubernetes-charts.banzaicloud.com/charts/canary-operator-${CANARY_OPERATOR_CHART_VERSION}.tgz | tar -zxv --directory ${CHARTS_DIR}/ -f -
retVal=$?
if [ $retVal -ne 0 ]; then
    exit $retVal
fi

curl -s https://kubernetes-charts.banzaicloud.com/charts/backyards-${BACKYARDS_CHART_VERSION}.tgz | tar -zxv --directory ${CHARTS_DIR}/ -f -
retVal=$?
if [ $retVal -ne 0 ]; then
    exit $retVal
fi
