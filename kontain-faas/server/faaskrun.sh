#!/bin/sh
#
# Create a simple bundle for a faas function, then create the config.json for that
# bundle, and then start krun to create the container and start the container to have
# the faas function executed.
#
# $1 - the directory to build the bundle under
# $2 - the executable for the faas function
# $3 - the directory contain the request and response data files.  This dir is bind mounted into the container.
# $4 - the file containing input for the faas function, this is relative to $3 (the faas data dir)
# $5 - the file the faas function should put its output in, this is relative to $3 (the faas data dir)
# $6 - the id to assign the container running the faas, must be unique
#

if test $# -lt 5
then
  echo "Usage: faaskrun.sh container-base-dir faas-function-executable faas-data-dir faas-function-input-file faas-function-output-file containerid"
  exit 1
else
  ROOTDIR=$1/faas_$$
  FAASPROG=$2
  FAASPROG_BN=`basename $2`
  FAASDATADIR=$3
  FAASINPUT=$4
  FAASOUTPUT=$5
  CONTAINERID=$6
fi
ROOTFS=rootfs
# The faas server puts the request file here and expects the respone file to be here.
FUNCDATA=kontain

# build our simple faas bundle
mkdir -p $ROOTDIR/$ROOTFS/opt/kontain
mkdir -p $ROOTDIR/$ROOTFS/usr/bin
cp $FAASPROG $ROOTDIR/$ROOTFS/usr/bin

# Note that krun will bind mount km into the container it runs

# build config.json
cat <<EOF >$ROOTDIR/config.json
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
            "/usr/bin/$FAASPROG_BN"
        ],
        "env": [
            "PATH=/usr/bin",
            "TERM=xterm"
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
            "source": "/$FAASDATADIR",
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
$KRUN run --no-new-keyring --bundle=$ROOTDIR $CONTAINERID
echo "krun returned $?"
echo "The contents of $FAASDATADIR/$FAASOUTPUT are:"
cat $FAASDATADIR/$FAASOUTPUT

# Remove the evidence
rm -fr $ROOTDIR
