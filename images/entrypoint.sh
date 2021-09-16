#!/bin/sh
set -e

CNDP_DIR="/cndp"
DP_BIN="cndp-dp"
DP_CONFIG_FILE=$CNDP_DIR/"config/config.json"
CNI_BIN="cndp"
CNI_BIN_DIR="/opt/cni/bin"

cp -f $CNDP_DIR/$CNI_BIN $CNI_BIN_DIR/$CNI_BIN
exec $CNDP_DIR/$DP_BIN -config $DP_CONFIG_FILE
