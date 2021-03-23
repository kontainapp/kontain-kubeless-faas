package main

import (
	"io"
	"os"
	"path"
	"os/exec"
	"fmt"
	"strings"
	"time"
)

func kontainApiCopyFile(srcPath string, dstPath string) error {
	src, err := os.OpenFile(srcPath, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(dstPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	return nil
}

func createKontainFunctionRootImage(rootPath string, containerId string, execPathName string, configPathName string) (string, error) {
	containerRootDir := rootPath + "/" + containerId + "/"
	imageRootDir := containerRootDir + "rootfs" + "/"

	kontainPath := imageRootDir + "/opt/kontain/"
	err := os.MkdirAll(kontainPath, 0777)
	if err != nil {
		return containerRootDir, err
	}
	binPath := imageRootDir + "/usr/bin/"
	err = os.MkdirAll(binPath, 0777)
	if err != nil {
		return containerRootDir, err
	}
	execTarget := binPath + path.Base(execPathName)
	err = kontainApiCopyFile(execPathName, execTarget)
	if err != nil {
		return containerRootDir, err
	}
	configTarget := containerRootDir + "/config.json"
	err = kontainApiCopyFile(configPathName, configTarget)
	if err != nil {
		return containerRootDir, err
	}
	return containerRootDir, nil
}

func cleanupKontainFunctionRootImage(containerRootDir string) {
	os.RemoveAll(containerRootDir)
}

func printDuration(tag string, start time.Time) {
	t := time.Now()
	delta := t.Sub(start)
	fmt.Printf("%s took %f seconds\n", tag, delta.Seconds());
}

/*
 * Create a config.json file in the passed bundle directory.
 * We substitute values into our model config.json string and then create the file and
 * write our complated model to it.  On failure any config.json created will be removed
 * before returning.
 * Returns:
 *   nil - success
 *   != nil - something failed, the returned error might be helpful
 */
func createConfigJson(configPath string, faasName string, faasDataDir string, requestFile string, responseFile string)  error {
	configJsonModel := `
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
            "/kontain/$FAASINPUT$",
            "--output-data",
            "/kontain/$FAASOUTPUT$",
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
	start := time.Now()

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
	printDuration("build config.json", start)
	return nil
}

/*
 * Given the base directory where container images and runtime bundles are kept,
 * and the name of the desired function, pull the needed function container and
 * setup the runtime bundle for that function.
 * Then return the path to the base of the runtime bundle.  This base of the runtime
 * bundle will contain the directory rootfs which will contain the container's files.
 * rootfs is supplied to krun via the config.json that some other function will need
 * to build.
 * If there is a failure, the error return value will be non-nil and anything this function
 * created before the failure will be removed.
 * Until we start reusing the runtime bundles the caller must delete the runtime bundle
 * once it is done using it.
 */
func getFunctionBundle(baseDir string, faasName string, containerId string) (string, error) {
	functionImageDir := baseDir + "/function_container_images/" + faasName
	bundleDir := baseDir + "/" + faasName

	// If we have a copy of the function container, use it
	_, err := os.Stat(functionImageDir)
	if err == nil {
		fmt.Printf("Using existing bundle for %s\n", faasName)
		return bundleDir, nil
	}

	fmt.Printf("Creating bundle for %s\n", faasName)

	start := time.Now()
	// make a directory for the container image
	err = os.MkdirAll(functionImageDir, 0755)
	if err != nil {
		return "", err
	}
	// pull the container image using skopeo
	dockerHost := os.Getenv("DOCKER_HOST")
	dockerCertPath := os.Getenv("DOCKER_CERT_PATH")
	execCmd := exec.Command("/usr/bin/skopeo", "copy", "--src-daemon-host", dockerHost, "--src-cert-dir", dockerCertPath, 
			"docker-daemon:" + faasName + ":latest",
			"oci:" + functionImageDir + ":latest");
	output, err := execCmd.CombinedOutput()
	fmt.Printf("Output of %s:\n%s\n==============\n", execCmd.String(), output)
	if err != nil {
		os.RemoveAll(functionImageDir)
		return "", err
	}
	printDuration("get container", start);


	// Make a directory for the runtime bundle.
	start = time.Now()
	err = os.MkdirAll(bundleDir, 0755)
	if err != nil {
		os.RemoveAll(functionImageDir)
		return "", err
	}
	// build the runtime bundle using oci-image-tool
	execCmd = exec.Command("/usr/bin/oci-image-tool", "unpack", "--ref", "name=latest", functionImageDir, bundleDir + "/rootfs")
	output, err = execCmd.CombinedOutput()
	fmt.Printf("Output of %s:\n%s\n==============\n", execCmd.String(), output)
	if err != nil {
		os.RemoveAll(functionImageDir)
		os.RemoveAll(bundleDir);
		return "", err
	}
	printDuration("unpack runtime bundle", start)

	return bundleDir, nil
}
