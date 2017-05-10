#!/usr/bin/env bash

set -e

mkdir -p ${PWD}/src
cp -a ./bosh-cpi-release/* ${PWD}/src
export GOPATH=${PWD}/src

pushd ${PWD}/src
  bin/build-amd64
  cp version out/softlayer_cpi promote/bosh-softlayer-cpi-patch
popd

apply_script='apply.sh'
cat > "${apply_script}"<<EOF
#!/bin/bash

TARGET_DIR="/var/vcap/packages/bosh_softlayer_cpi"
TARGET_DIR_BIN=$TARGET_DIR/bin

# check if the eCPI current version is different with target version
if [ ! -f $TARGET_DIR/version ]; then
	CURRENT_VERSION="unknown"
else
	CURRENT_VERSION=`cat $TARGET_DIR/version`
fi
TARGET_VERSION=`cat version`

if [ "$TARGET_VERSION" == "$CURRENT_VERSION" ]; then
	echo "The current version of eCPI is the same as target version $TARGET_VERSION. Won't do any upgrade. Exit directly."
	exit 1
else
	echo "Current version of eCPI is $CURRENT_VERSION. Will upgrade to version $TARGET_VERSION ..."
fi

# backup the current version
if [ -f $TARGET_DIR/version ]; then
	mv $TARGET_DIR/version $TARGET_DIR/version_$CURRENT_VERSION
fi
mv $TARGET_DIR_BIN/softlayer_cpi $TARGET_DIR_BIN/softlayer_cpi_$CURRENT_VERSION

# copy target bosh_softlayer_cpi and version files to the corresponding dir
cp softlayer_cpi $TARGET_DIR_BIN
cp version $TARGET_DIR
chmod 755 $TARGET_DIR_BIN/softlayer_cpi
chmod 644 $TARGET_DIR/version

# simply verify new softlayer_cpi
$TARGET_DIR_BIN/softlayer_cpi -version
if [ $? == 0 ]; then
	echo "Simply verification of $TARGET_DIR_BIN/softlayer_cpi passed!"
else
	echo "$TARGET_DIR_BIN/softlayer_cpi can't run! Please check the patch!"
	exit 1
fi
EOF

cp ${apply_script} promote/bosh-softlayer-cpi-patch
cd promote/bosh-softlayer-cpi-patch
version_number=`cat version`
tar -zcvf bosh-softlayer-cpi-patch-${version_number}.tgz ${apply_script} version softlayer_cpi
