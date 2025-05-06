package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
)

func generateTimeBasedVersion() string {
	// vYY.MM.DD-eph.SECONDS_SINCE_MIDNIGHT
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	secondsSinceMidnight := int(now.Sub(midnight).Seconds())
	return "v" + fmt.Sprintf("%02d.%02d.%02d", now.Year()%100, now.Month(), now.Day()) + "-eph." + strconv.Itoa(secondsSinceMidnight)
}

func buildDockerImage(cmdDepth ...string) error {
	version := generateTimeBasedVersion()
	appName := makeCmdDepthToName(cmdDepth...)
	imageTag := globalConfig.Namespace + "/" + appName + ":" + version
	dockerFilePath := filepath.Join(RootPackageCmd, filepath.Join(cmdDepth...), "Dockerfile")
	cmd := exec.Command("docker", "build", "-t", imageTag, "-f", dockerFilePath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	originImageTagForLocal := globalConfig.LocalRegistry + "/" + imageTag
	log.Printf("Building docker image for local image: %s", originImageTagForLocal)
	sp := strings.Split(globalConfig.LocalRegistry, ".")
	imageTagForLocal := sp[len(sp)-1] + "/" + imageTag
	log.Printf("Build local image tag: %s", imageTagForLocal)

	cmd = exec.Command("docker", "tag", imageTag, imageTagForLocal)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("docker", "push", imageTagForLocal)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Printf("pushed image to local registry: %s", color.New(color.Bold).Sprint(originImageTagForLocal))

	imageTagForRemote := globalConfig.RemoteRegistry + "/" + imageTag

	log.Printf("Building docker image for remote image %s", color.New(color.Bold).Sprint(imageTagForLocal))

	cmd = exec.Command("docker", "tag", imageTag, imageTagForRemote)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Printf("push image %s yourself, I cannot assist.", color.New(color.Bold).Sprint(imageTagForRemote))

	return nil
}
