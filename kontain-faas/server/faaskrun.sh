#!/bin/sh
#
# Create a simple bundle for a faas function, then create the config.json for that
# bundle, and then start krun to create the container and start the container to have
# the faas function executed.
#
# $1 - the directory to build the bundle under
# $2 - the faas function name
# $3 - the directory contain the request and response data files.  This dir is bind mounted into the container. Must be a relative path.
# $4 - the file containing input for the faas function, this is relative to $3 (the faas data dir)
# $5 - the file the faas function should put its output in, this is relative to $3 (the faas data dir)
# $6 - the id to assign the container running the faas, must be unique
#
# This script creates the following directories and files under the directory passed as $1
#  $1/function_container_images/function1
#                              /function2
#                              /functionN
#                              /$2
#  $1/$2_$6/bundledir/config.json
#                    /rootfs/.....
#
# skopeo copy --src-daemon-host $DOCKER_HOST --src-cert-dir $DOCKER_CERT_PATH docker-daemon:test_func_data_with_hc:latest oci:blort:latest
# oci-image-tool unpack  --ref name=latest blort blatt
SKOPEO=skopeo
#OCI_IMAGE_TOOL=/home/paulp/go/src/github.com/opencontainers/image-tools/oci-image-tool
OCI_IMAGE_TOOL=oci-image-tool

if test $# -lt 5
then
  echo "Usage: faaskrun.sh fass-work-dir faas-function-name faas-data-dir faas-function-input-file faas-function-output-file containerid"
  exit 1
else
  FAASWORKDIR=$1
  FAASFUNC=$2
  FAASDATADIR=`pwd`/$3
  FAASINPUT=$4
  FAASOUTPUT=$5
  CONTAINERID=$6
fi
ROOTFS=rootfs
CONTAINERDIR=function_contain_images

# See if we have the container image for this faas function.
# Don't have it, use skopeo to get it.
mkdir -p $FAASWORKDIR/$CONTAINERDIR
if [ ! -d $FAASWORKDIR/$CONTAINERDIR/$FAASFUNC ]
then
    mkdir -p $FAASWORKDIR/$CONTAINERDIR/$FAASFUNC
    pushd $FAASWORKDIR/$CONTAINERDIR
    $SKOPEO copy --src-daemon-host $DOCKER_HOST --src-cert-dir $DOCKER_CERT_PATH docker-daemon:$FAASFUNC:latest oci:$FAASFUNC:latest
    RC=$?
    popd
    if [ $RC != 0 ]
    then
        rm -fr $FAASWORKDIR/$CONTAINERDIR/$FAASFUNC
        exit 1
    fi
fi

# Create the runtime bundle for the function
mkdir -p $FAASWORKDIR/${FAASFUNC}_$CONTAINERID
$OCI_IMAGE_TOOL unpack --ref name=latest $FAASWORKDIR/$CONTAINERDIR/$FAASFUNC $FAASWORKDIR/${FAASFUNC}_$CONTAINERID/rootfs
RC=$?
if [ $RC -ne 0 ]
then
    rm -fr $FAASWORKDIR/$FAASFUNC_$CONTAINERID
    exit 1
fi

# build config.json
cat <<EOF >$FAASWORKDIR/${FAASFUNC}_$CONTAINERID/config.json
{
    "ociVersion": "1.0.0",
    "process": {
        "user": {
            "uid": 0,
            "gid": 0
        },
        "terminal": false,
        "args": [
            "/opt/kontain/bin/km",
            "--input-data",
            "/kontain/$FAASINPUT",
            "--output-data",
            "/kontain/$FAASOUTPUT",
            "/usr/bin/$FAASFUNC.km"
        ],
        "env": [
            "PATH=/usr/bin",
            "TERM=xterm",
            "DOCKER_CERT_PATH=/home/paulp/.minikube/certs",
            "DOCKER_HOST=tcp://192.168.49.2:2376"
        ],
        "cwd": "/",
        "noNewPrivileges": true
    },
    "root": {
        "path": "$ROOTFS"
    },
    "mounts": [
        {
            "destination": "/proc",
            "type": "proc"
        },
        {
            "destination": "/sys",
            "type": "sysfs",
            "source": "sysfs",
            "options": [
                "nosuid",
                "noexec",
                "nodev",
                "ro"
            ]
        },
        {
            "destination": "/sys/fs/cgroup",
            "type": "cgroup",
            "source": "cgroup",
            "options": [
                "nosuid",
                "noexec",
                "nodev",
                "relatime",
                "rw"
            ]
        },
        {
            "destination": "/dev",
            "type": "tmpfs",
            "source": "tmpfs",
            "options": [
                "nosuid",
                "strictatime",
                "mode=755",
                "size=65536k"
            ]
        },
        {
            "destination": "/dev/pts",
            "type": "devpts",
            "source": "devpts",
            "options": [
                "nosuid",
                "noexec",
                "newinstance",
                "ptmxmode=0666",
                "mode=0620"
            ]
        },
        {
            "destination": "/dev/shm",
            "type": "tmpfs",
            "source": "shm",
            "options": [
                "nosuid",
                "noexec",
                "nodev",
                "mode=1777",
                "size=65536k"
            ]
        },
        {
            "destination": "/dev/mqueue",
            "type": "mqueue",
            "source": "mqueue",
            "options": [
                "nosuid",
                "noexec",
                "nodev"
            ]
        },
        {
            "destination": "/kontain",
            "type": "none",
            "source": "$FAASDATADIR",
            "options": ["bind", "rw"]
        }
    ],
    "linux": {
        "rootfsPropagation": "rprivate",
        "namespaces": [
            {
                "type": "mount"
            },
            {
                "type": "pid"
            },
            {
                "type": "user"
            },
            {
                "type": "ipc"
            },
            {
                "type": "cgroup"
            }
        ]
    }
}
EOF

# run the container
KRUN=/opt/kontain/bin/krun
$KRUN run --no-new-keyring --bundle=$FAASWORKDIR/${FAASFUNC}_$CONTAINERID $CONTAINERID
RC=$?
echo "krun returned $RC"
echo "The contents of $FAASDATADIR/$FAASOUTPUT are:"
cat $FAASDATADIR/$FAASOUTPUT

# Remove the evidence
if [ $RC -eq 0 ]
then
    rm -fr $FAASWORKDIR/${FAASFUNC}_$CONTAINERID
fi
