// Copyright © 2021 Kontain Inc. All rights reserved.
//
// Kontain Inc CONFIDENTIAL
//
// This file includes unpublished proprietary source code of Kontain Inc. The
// copyright notice above does not evidence any actual or intended publication
// of such source code. Disclosure of this source code or any related
// proprietary information is strictly prohibited without the express written
// permission of Kontain Inc.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func timeTaken(msg string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s took %s\n", msg, time.Since(start))
	}
}

// Given the base directory where container images and runtime bundles are kept,
// and the name of the desired function, pull the needed function container and
// setup the runtime bundle for that function.
// This base of the runtime
// bundle will contain the directory rootfs which will contain the container's files.
// rootfs is supplied to krun via the config.json that some other function will need
// to build.
// If there is a failure, the error return value will be non-nil and anything this function
// created before the failure will be removed.
// Until we start reusing the runtime bundles the caller must delete the runtime bundle
// once it is done using it.
func getFunctionBundle(faasName string, functionImageDir string, bundleDir string) error {

	// If we have a copy of the function container, use it
	_, err := os.Stat(functionImageDir)
	if err == nil {
		fmt.Printf("Using existing bundle for %s\n", faasName)
		return nil
	}

	err = fetchBundle(functionImageDir, faasName)
	if err != nil {
		return err
	}

	err = makeBundleDir(bundleDir, functionImageDir, faasName)
	if err != nil {
		os.RemoveAll(functionImageDir)
		return err
	}

	return nil
}

func fetchBundle(functionImageDir string, faasName string) error {
	defer timeTaken("fetchBundle")()

	fmt.Printf("Fetch bundle for %s\n", faasName)

	// make a directory for the container image
	err := os.MkdirAll(functionImageDir, 0755)
	if err != nil {
		return err
	}

	// pull the container image using skopeo
	dockerHost := os.Getenv("DOCKER_HOST")
	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	execCmd := exec.Command("/usr/bin/skopeo", "copy",
		"--src-daemon-host", dockerHost, "--src-cert-dir", dockerCertPath,
		"docker-daemon:"+faasName+":latest", "oci:"+functionImageDir+":latest")
	output, err := execCmd.CombinedOutput()
	fmt.Printf("Output of %s:\n%s\n==============\n", execCmd.String(), output)
	if err != nil {
		os.RemoveAll(functionImageDir)
		return err
	}

	return nil
}

func makeBundleDir(bundleDir string, functionImageDir string, faasName string) error {
	defer timeTaken("makeBundleDir")()

	fmt.Printf("Create function image for %s\n", faasName)

	// Make a directory for the runtime bundle.
	err := os.MkdirAll(bundleDir, 0755)
	if err != nil {
		return err
	}

	// build the runtime bundle using oci-image-tool
	execCmd := exec.Command("/usr/bin/oci-image-tool", "unpack",
		"--ref", "name=latest", functionImageDir, bundleDir+"/rootfs")
	output, err := execCmd.CombinedOutput()
	fmt.Printf("Output of %s:\n%s\n==============\n", execCmd.String(), output)
	if err != nil {
		os.RemoveAll(bundleDir)
		return err
	}

	return nil
}

// Create a config.json file in the passed bundle directory.
// We substitute values into our model config.json string and then create the file and
// write our complated model to it.  On failure any config.json created will be removed
// before returning.
// Returns:
//   nil - success
//   != nil - something failed, the returned error might be helpful
func createConfigJson(configPath string, faasName string,
	faasDataDir string, requestFile string, responseFile string) error {
	defer timeTaken("createConfigJson")()

	// Substitute the function and instance specific values into our config.json pattern string.
	configJson := strings.ReplaceAll(configJsonModel, "$FAASINPUT$", requestFile)
	configJson = strings.ReplaceAll(configJson, "$FAASOUTPUT$", responseFile)
	configJson = strings.ReplaceAll(configJson, "$FAASFUNC$", faasName)
	configJson = strings.ReplaceAll(configJson, "$FAASDATADIR$", faasDataDir)

	// Write the config.json file to the bundle directory.
	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(configJson)
	if err != nil {
		os.Remove(configPath)
		return err
	}

	return nil
}

const (
	configJsonModel string = `
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
            "$FAASINPUT$",
            "--output-data",
            "$FAASOUTPUT$",
            "/usr/bin/$FAASFUNC$.km"
        ],
        "env": [
            "PATH=/usr/bin",
            "TERM=xterm"
        ],
        "cwd": "/",
        "noNewPrivileges": true
    },
    "root": {
        "path": "rootfs",
        "readonly": true
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
            "source": "$FAASDATADIR$",
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
`
)
