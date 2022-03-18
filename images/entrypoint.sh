#!/bin/sh
set -e

BINS_DIR="/afxdp"
DP_BIN="afxdp-dp"
DP_CONFIG_FILE=$BINS_DIR/"config/config.json"
CNI_BIN="afxdp"
CNI_BIN_DIR="/opt/cni/bin"

cp -f $BINS_DIR/$CNI_BIN $CNI_BIN_DIR/$CNI_BIN
exec $BINS_DIR/$DP_BIN -config $DP_CONFIG_FILE
