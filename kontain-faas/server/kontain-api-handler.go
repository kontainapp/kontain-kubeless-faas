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

var path_name string = "/kontain/"

func request_file_name(faas_name string) string {
	return path_name + "/" + faas_name + ".request"
}
func response_file_name(faas_name string) string {
	return path_name + "/" + faas_name + ".response"
}
func exec_file_name(faas_name string) string {
	return path_name + "/" + faas_name
}

func GetCallFunction(url string) (string, error) {
	comp := strings.Split(url, "/")
	if len(comp) == 0 {
		return "", errors.New("Invalid URL")
	}
	fp := exec_file_name(comp[1])
	finfo, err := os.Stat(fp)
	if err != nil {
		return comp[1], err
	}
	if finfo.Mode().Perm()&0111 != 0111 {
		return comp[1], errors.New("Invalid permissions on executable")
	}
	return comp[1], nil
}

func ApiHandlerExecCallFunction(faas_name string) error {
	fp := exec_file_name(faas_name)
	exec_cmd := exec.Command(fp)
	err := exec_cmd.Run()
	return err
}

func ApiHandlerWriteRequest(faas_name string, method string, url string, header http.Header, data []byte) error {
	fn := request_file_name(faas_name)
	fd, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0777)
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

	fmt.Fprintf(fd, "DATA: %s\n", base64.StdEncoding.EncodeToString(data))

	return nil
}

func ApiHandlerReadRequest(faas_name string) (string, string, []byte, error) {
	var method string
	var url string
	var dec_data []byte
	header := make(map[string][]string)

	fn := request_file_name(faas_name)
	fd, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return method, url, dec_data, err
	}
	defer fd.Close()

	sc := bufio.NewScanner(fd)
	for sc.Scan() {
		comp := strings.Split(sc.Text(), ":")
		if len(comp) != 2 {
			return method, url, dec_data, errors.New("Corrupted input")
		}
		comp1 := strings.TrimSpace(comp[0])
		comp2 := strings.TrimSpace(comp[1])
		switch comp1 {
		case "METHOD":
			method = comp2
		case "URL":
			url_byte, err := base64.StdEncoding.DecodeString(comp2)
			if err != nil {
				return method, url, dec_data, err
			}
			url = string(url_byte)
		case "HEADER":
			hcomp := strings.Split(comp2, ",")
			if len(hcomp) != 2 {
				return method, url, dec_data, errors.New("Corrupted input in HEADER")
			}
			var k string
			var v []string
			for i, j := range hcomp {
				j_dec, err := base64.StdEncoding.DecodeString(strings.TrimSpace(j))
				if err != nil {
					return method, url, dec_data, err
				}
				if i == 0 {
					k = string(j_dec)
				} else {
					v = append(v, string(j_dec))
				}
			}
			header[k] = v
		case "DATA":
			dec_data, err = base64.StdEncoding.DecodeString(comp2)
			if err != nil {
				return method, url, dec_data, err
			}
		}
	}
	return method, url, dec_data, nil
}

func ApiHandlerWriteResponse(faas_name string, status_code int, data []byte) error {
	fn := response_file_name(faas_name)
	fd, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer fd.Close()

	fmt.Fprintf(fd, "STATUSCODE: %d\n", http.StatusOK)
	fmt.Fprintf(fd, "DATA: %s\n", base64.StdEncoding.EncodeToString(data))

	return nil
}

func ApiHandlerReadResponse(faas_name string) (int, []byte, error) {
	code := http.StatusNotFound
	dec_data := []byte("")

	fn := response_file_name(faas_name)
	fd, err := os.OpenFile(fn, os.O_RDONLY, 0)
	if err != nil {
		return code, dec_data, err
	}
	defer fd.Close()

	sc := bufio.NewScanner(fd)
	for sc.Scan() {
		comp := strings.Split(sc.Text(), ":")
		comp1 := strings.TrimSpace(comp[0])
		comp2 := strings.TrimSpace(comp[1])
		switch comp1 {
		case "STATUSCODE":
			code, _ = strconv.Atoi(strings.TrimSpace(comp2))
		case "DATA":
			dec_data, _ = base64.StdEncoding.DecodeString(comp2)
		}
	}

	return code, dec_data, nil
}
