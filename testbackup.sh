#! /usr/bin/env bash
#
set -x
env
ls -la ${BACKIVE_TO}
echo "This is a test..."
echo "mount/to $BACKIVE_TO"
echo "mount/from $BACKIVE_FROM"
echo "mount $BACKIVE_MOUNT"
cp -Rv ${BACKIVE_FROM}/* ${BACKIVE_TO}/
ls -la ${BACKIVE_MOUNT}
ls -la ${BACKIVE_TO}
