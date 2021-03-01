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
)

var pathName string = "/kontain/"

func requestFileName(faasName string) string {
	return pathName + "/" + faasName + ".request"
}
func responseFileName(faasName string) string {
	return pathName + "/" + faasName + ".response"
}
func execFileName(faasName string) string {
	return pathName + "/" + faasName
}

func GetCallFunction(url string) (string, error) {
	comp := strings.Split(url, "/")
	if len(comp) == 0 {
		return "", errors.New("Invalid URL")
	}
	fp := execFileName(comp[1])
	finfo, err := os.Stat(fp)
	if err != nil {
		return comp[1], err
	}
	if finfo.Mode().Perm()&0111 != 0111 {
		return comp[1], errors.New("Invalid permissions on executable")
	}
	return comp[1], nil
}

func ApiHandlerExecCallFunction(faasName string) error {
	fp := execFileName(faasName)
	execCmd := exec.Command(fp)
	err := execCmd.Run()
	return err
}

func ApiHandlerWriteRequest(faasName string, method string, url string, header http.Header, data string) error {
	fn := requestFileName(faasName)
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

func ApiHandlerReadRequest(faasName string) (string, string, []byte, error) {
	var method string
	var url string
	var decData []byte
	header := make(map[string][]string)

	fn := requestFileName(faasName)
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

func ApiHandlerWriteResponse(faasName string, statusCode int, data []byte) error {
	fn := responseFileName(faasName)
	fd, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return err
	}
	defer fd.Close()

	fmt.Fprintf(fd, "STATUSCODE: %d\n", statusCode)
	fmt.Fprintf(fd, "DATA: %s\n", base64.StdEncoding.EncodeToString(data))

	return nil
}

func ApiHandlerReadResponse(faasName string) (int, []byte, error) {
	code := http.StatusNotFound
	decData := []byte("")

	fn := responseFileName(faasName)
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
