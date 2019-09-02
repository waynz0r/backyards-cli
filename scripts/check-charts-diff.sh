#!/bin/bash
CHARTS_DIR=${1:-/tmp/.charts-diff}
ORIG_CHARTS_DIR=${2:-assets/charts}
DIR=$(dirname $0)

mkdir -p $CHARTS_DIR
retVal=$?
if [ $retVal -ne 0 ]; then
    exit $retVal
fi

$DIR/download-charts.sh $CHARTS_DIR
retVal=$?
if [ $retVal -ne 0 ]; then
    exit $retVal
fi

diff -urN $CHARTS_DIR $ORIG_CHARTS_DIR
retVal=$?
if [ $retVal -ne 0 ]; then
    rm -rf $CHARTS_DIR
    exit $retVal
fi

echo "no difference - ok!"

rm -rf $CHARTS_DIR
