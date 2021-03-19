package main

import (
	"io"
	"os"
	"path"
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
