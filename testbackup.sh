#! /usr/bin/env bash
#
set -x
ls -la ${BACKIVE_TO}
echo "This is a test..."
echo "mount/to ${BACKIVE_TO}"
cp -Rv ${BACKIVE_FROM}/* ${BACKIVE_TO}/
ls -la ${BACKIVE_MOUNT}
ls -la ${BACKIVE_TO}
