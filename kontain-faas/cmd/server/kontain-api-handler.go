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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
)

type FaasCall struct {
	Namespace    string
	Function     string
	Id           string
	BundlePath   string
	InstancePath string
	RequestPath  string
	ReplyPath    string
	ConfigPath   string
	StdErrOut    string
}

func runtimeBundlePath(namespace string, function string) string {
	return "/kontain/oci-runtime-bundles/" + namespace + "/" + function
}

func runtimeInstancePath(namespace string, function string) string {
	return "/kontain/runtime-instances/" + namespace + "/" + function
}

func FaasApiGetFunctionInstance(url string) *FaasCall {

	comp := strings.Split(url, "/")
	if len(comp) < 3 {
		return nil
	}
	namespace := comp[1]
	function := comp[2]

	path := runtimeBundlePath(namespace, function)
	fi, err := os.Stat(path)
	if err != nil || !fi.IsDir() {
		return nil
	}
	id := uuid.NewString()
	instancePath := runtimeInstancePath(namespace, function) + "/" + id
	ret := &FaasCall{
		Namespace:    namespace,
		Function:     function,
		Id:           id,
		BundlePath:   path,
		InstancePath: instancePath,
		ConfigPath:   instancePath + "/config.json",
		RequestPath:  instancePath + "/request",
		ReplyPath:    instancePath + "/reply",
		StdErrOut:    instancePath + "/out+err",
	}
	err = os.MkdirAll(ret.InstancePath, 0755)
	if err != nil {
		return nil
	}

	return ret
}

func (f *FaasCall) HandlerExecCallFunction() error {

	if err := f.CreateConfig(); err != nil {
		return err
	}

	// Make sure reply file exists
	{
		file, err := os.Create(f.ReplyPath)
		if err != nil {
			return err
		}
		file.Close()
	}
	execCmd := exec.Command("/opt/kontain/bin/krun", "run", "--no-new-keyring", "--config="+f.ConfigPath,
		"--bundle="+f.BundlePath+"/rootfs", f.Id)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: execCmd failed err=%v outut=%s\n", err, output)
		return err
	}
	return nil
}

type FunctionRequest struct {
	Method  string
	Url     string
	Headers map[string][]string
	Data    string
}

type FunctionReply struct {
	Status int
	Data   []byte
}

func (f *FaasCall) HandlerWriteRequest(method string, url string, header http.Header, data string) error {
	req, err := json.Marshal(&FunctionRequest{
		Method: method,
		Url:    url,
		Data:   base64.StdEncoding.EncodeToString([]byte(data)),
	})
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(f.RequestPath, req, 0444)
	if err != nil {
		return err
	}

	return nil
}

func (f *FaasCall) HandlerReadRequest() (string, string, []byte, error) {
	var method string
	var url string
	var decData []byte
	return method, url, decData, nil
	/*
		header := make(map[string][]string)

		fn := requestPathName(faasName, id)
		fd, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0777)
		if err != nil {
			return method, url, decData, err
		}
		defer fd.Close()

		sc := bufio.NewScanner(fd)
		for sc.Scan() {
			comp := strings.Split(sc.Text(), ":")
			if len(comp) != 2 {
				return method, url, decData, errors.New("Corrupted input")
			}
			comp1 := strings.TrimSpace(comp[0])
			comp2 := strings.TrimSpace(comp[1])
			switch comp1 {
			case "METHOD":
				method = comp2
			case "URL":
				urlByte, err := base64.StdEncoding.DecodeString(comp2)
				if err != nil {
					return method, url, decData, err
				}
				url = string(urlByte)
			case "HEADER":
				hcomp := strings.Split(comp2, ",")
				if len(hcomp) != 2 {
					return method, url, decData, errors.New("Corrupted input in HEADER")
				}
				var k string
				var v []string
				for i, j := range hcomp {
					jDec, err := base64.StdEncoding.DecodeString(strings.TrimSpace(j))
					if err != nil {
						return method, url, decData, err
					}
					if i == 0 {
						k = string(jDec)
					} else {
						v = append(v, string(jDec))
					}
				}
				header[k] = v
			case "DATA":
				decData, err = base64.StdEncoding.DecodeString(comp2)
				if err != nil {
					return method, url, decData, err
				}
			}
		}
		return method, url, decData, nil
	*/
}

func ApiHandlerWriteResponse(faasName string, id string, statusCode int, data []byte) error {
	/*
		fn := responsePathName(faasName, id)
		fd, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			return err
		}
		defer fd.Close()

		fmt.Fprintf(fd, "STATUSCODE: %d\n", statusCode)
		fmt.Fprintf(fd, "DATA: %s\n", base64.StdEncoding.EncodeToString(data))
	*/

	return nil
}

func (f *FaasCall) HandlerReadResponse() (int, []byte, error) {

	data, err := ioutil.ReadFile(f.ReplyPath)
	if err != nil {
		return http.StatusNotFound, data, err
	}

	reply := FunctionReply{}
	err = json.Unmarshal([]byte(data), &reply)
	if err != nil {
		return http.StatusNotFound, data, err
	}
	return reply.Status, reply.Data, nil
}

func (f *FaasCall) HandlerCleanFiles() {
	/*
		if err := os.RemoveAll(f.InstancePath); err != nil {
			// TODO: bark
		}
	*/
}

// OCI Image Configurationa - blob, pointed to by OCIImageManifest.Config
type OCIImageConfiguration struct {
	Created      string             `json:"created,omitempty"`
	Author       string             `json:"author,omitempty"`
	Architecture string             `json:"architecture"`
	Os           string             `json:"os"`
	Config       OCIImageExecConfig `json:"config"`
}

//
type OCIImageExecConfig struct {
	User         string              `json:"User,omitempty"`
	ExposedPorts map[string]struct{} `json:"ExposedPorts,omitempty"`
	Env          []string            `json:"Env,omitempty"`
	Entrypoint   []string            `json:"ENtrypoint,omitempty"`
	Cmd          []string            `Entrypoint:"Cmd,omitempty"`
	Volumes      map[string]struct{} `json:"Volumesexposedports,omitempty"`
	WorkingDir   string              `json:"WorkingDir,omitempty"`
	Labels       map[string]string   `jsoon:"Labels,omitempty"`
	StopSignal   string              `json:"StopSignal,omitempty"`
}

func (f *FaasCall) CreateConfig() error {
	data, err := ioutil.ReadFile(f.BundlePath + "/oci-config.json")
	if err != nil {
		return err
	}
	var imageConfiguration OCIImageConfiguration
	json.Unmarshal(data, &imageConfiguration)

	cmd := ""
	for _, c := range imageConfiguration.Config.Entrypoint {
		if len(cmd) > 0 {
			cmd = cmd + ", "
		}
		cmd = cmd + "\"" + c + "\""
	}
	for _, c := range imageConfiguration.Config.Cmd {
		if len(cmd) > 0 {
			cmd = cmd + ", "
		}
		cmd = cmd + "\"" + c + "\""
	}

	env := ""
	for _, c := range imageConfiguration.Config.Env {
		if len(env) > 0 {
			env = env + ", "
		}
		env = env + "\"" + c + "\""
	}

	// Substitute the function and instance specific values into our config.json pattern string.
	configJson := strings.ReplaceAll(configJsonTemplate, "$FAASINPUT$", f.RequestPath)
	configJson = strings.ReplaceAll(configJson, "$FAASOUTPUT$", f.ReplyPath)
	configJson = strings.ReplaceAll(configJson, "$FAASFUNC$", f.Function)
	configJson = strings.ReplaceAll(configJson, "$FAASDATADIR$", f.InstancePath)
	configJson = strings.ReplaceAll(configJson, "$FAASSTDERROUT$", f.StdErrOut)
	configJson = strings.ReplaceAll(configJson, "$FAASBUNDLEPATH$", f.BundlePath+"/rootfs")
	configJson = strings.ReplaceAll(configJson, "$FAASCMD$", cmd)
	configJson = strings.ReplaceAll(configJson, "$FAASENV$", env)

	err = ioutil.WriteFile(f.ConfigPath, []byte(configJson), 0444)
	if err != nil {
		return err
	}
	return nil
}

/*
 */
const (
	configJsonTemplate string = `
{
    "ociVersion": "1.0.0",
    "annotations": {
        "run.oci.hooks.stdout": "$FAASSTDERROUT$",
        "run.oci.hooks.stderr": "$FAASSTDERROUT$"
	},
    "process": {
        "user": {
            "uid": 0,
            "gid": 0
        },
        "terminal": false,
        "args": [
			"/opt/kontain/bin/km",
			"--coredump",
			"/kontain/kmcore",
			"--snapshot",
			"/kontain/kmsnap",
            "--input-data",
            "/.request",
            "--output-data",
            "/.reply",
            $FAASCMD$
        ],
        "env": [
            $FAASENV$
        ],
        "cwd": "/",
        "noNewPrivileges": true
    },
    "root": {
        "path": "$FAASBUNDLEPATH$",
        "readonly": true
    },
    "mounts": [
        {
            "destination": "/.request",
            "type": "none",
            "source": "$FAASINPUT$",
			"options": [ "bind" ]
        },
        {
            "destination": "/.reply",
            "type": "none",
            "source": "$FAASOUTPUT$",
            "options": [ "bind" ]
        },
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
