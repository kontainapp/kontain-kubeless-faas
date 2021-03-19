package main

import (
	"bufio"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync/atomic"
)

type KontainApi struct {
	SerialId uint64
}

var kontainApi KontainApi

func init() {
	kontainApi.SerialId = 1
}

func getNextSerialId() string {
	id := atomic.AddUint64(&kontainApi.SerialId, 1)
	//return fmt.Sprintf("%016x", id)
	_ = id
	return fmt.Sprintf("%016x", 0)
}

var pathName string = "/kontain/"

func requestFileName(faasName string, id string) string {
	return faasName + "-" + id + ".request"
}
func requestPathName(faasName string, id string) string {
	return pathName + "/" + requestFileName(faasName, id)
}
func responseFileName(faasName string, id string) string {
	return faasName + "-" + id + ".response"
}
func responsePathName(faasName string, id string) string {
	return pathName + "/" + responseFileName(faasName, id)
}

// Map faas function name to kontain payload
func execPathName(faasName string) string {
	return pathName + faasName + ".km"
}
func configPathName(faasName string) string {
	return pathName + faasName + ".json"
}

func GetCallFunction(url string) (string, error) {
	comp := strings.Split(url, "/")
	if len(comp) == 0 {
		return "", errors.New("Invalid URL")
	}
	fp := execPathName(comp[1])
	finfo, err := os.Stat(fp)
	if err != nil {
		return fp, err
	}
	if finfo.Mode().Perm()&0111 != 0111 {
		return comp[1], errors.New("Invalid permissions on executable")
	}
	return comp[1], nil
}

func ApiHandlerExecCallFunction(faasName string, id string) error {
	containerBaseDir := "run_faas_here"
	execPath := execPathName(faasName)
	configPath := configPathName(faasName)
	//rq := requestFileName(faasName, id)
	//rp := responseFileName(faasName, id)
	containerId := faasName + "-" + id

	cotainerRootDir, err := createKontainFunctionRootImage(pathName+"/"+containerBaseDir, containerId, execPath, configPath)
	defer cleanupKontainFunctionRootImage(cotainerRootDir)
	if err != nil {
		return err
	}

	execCmd := exec.Command("/opt/kontain/bin/krun", "run", "--no-new-keyring", "--bundle="+cotainerRootDir, containerId)
	err = execCmd.Run()
	return err
}

func ApiHandlerWriteRequest(faasName string, method string, url string, id string, header http.Header, data string) error {
	fn := requestPathName(faasName, id)
	fd, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer fd.Close()
	fmt.Fprintf(fd, "METHOD: %s\n", method)
	fmt.Fprintf(fd, "URL: %s\n", base64.StdEncoding.EncodeToString([]byte(url)))

	for k, v := range header {
		fmt.Fprintf(fd, "HEADER: %s",
			base64.StdEncoding.EncodeToString([]byte(k)))
		for _, j := range v {
			fmt.Fprintf(fd, " ,%s",
				base64.StdEncoding.EncodeToString([]byte(j)))
		}
		fmt.Fprintf(fd, "\n")
	}

	fmt.Fprintf(fd, "DATA: %s\n", base64.StdEncoding.EncodeToString([]byte(data)))

	return nil
}

func ApiHandlerReadRequest(faasName string, id string) (string, string, []byte, error) {
	var method string
	var url string
	var decData []byte
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
}

func ApiHandlerWriteResponse(faasName string, id string, statusCode int, data []byte) error {
	fn := responsePathName(faasName, id)
	fd, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer fd.Close()

	fmt.Fprintf(fd, "STATUSCODE: %d\n", statusCode)
	fmt.Fprintf(fd, "DATA: %s\n", base64.StdEncoding.EncodeToString(data))

	return nil
}

func ApiHandlerReadResponse(faasName string, id string) (int, []byte, error) {
	code := http.StatusNotFound
	decData := []byte("")

	fn := responsePathName(faasName, id)
	fd, err := os.OpenFile(fn, os.O_RDONLY, 0)
	if err != nil {
		return code, decData, err
	}
	defer fd.Close()

	sc := bufio.NewScanner(fd)
	for sc.Scan() {
		comp := strings.Split(sc.Text(), ":")
		comp1 := strings.TrimSpace(comp[0])
		comp2 := strings.TrimSpace(comp[1])
		switch comp1 {
		case "STATUSCODE":
			code, err = strconv.Atoi(strings.TrimSpace(comp2))
			if err != nil {
				return code, decData, err
			}
		case "DATA":
			decData, err = base64.StdEncoding.DecodeString(comp2)
			if err != nil {
				return code, decData, err
			}
		}
	}

	return code, decData, nil
}

func ApiHandlerCleanFiles(faasName string, id string) {
	reqFn := requestPathName(faasName, id)
	os.Remove(reqFn)
	resFn := responsePathName(faasName, id)
	os.Remove(resFn)
}
