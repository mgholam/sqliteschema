package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// must have powershell on windows
const winZigPath = "c:/zig/zig.exe"
const outputDir = "output/"

func main() {
	incrementBuildNumber()
	os.Mkdir(outputDir, 0777)
	name := "sqliteschema"
	build(name+"-linux", "linux", "amd64", "x86_64-linux-musl", 1)
	// if need to strip
	// exec.Command("strip", outputDir+name+"-linux").Run()

	//	copyFile(outputDir+name+"-linux", "deploy/"+name)

	build(name+"-w64.exe", "windows", "amd64", "x86_64-windows-gnu", 1)
	build(name+"-piz", "linux", "arm GOARM=5", "arm-linux-musleabihf -march=arm1176jzf_s", 1)

	// build(name+"-android", "linux", "arm64", "aarch64-linux-musl", 1)
	// build(name+"-mac", "darwin", "amd64", "aarch64-macos-gnu", 1)
}

func incrementBuildNumber() {
	b, _ := os.ReadFile("version.go")
	reg := regexp.MustCompile("BUILD = ([0-9]*)")
	s := reg.FindStringSubmatch(string(b))
	bn, _ := strconv.Atoi(s[1])
	bn++

	t := strings.Replace(string(b), s[1], fmt.Sprintf("%d", bn), 1)
	os.WriteFile("version.go", []byte(t), 0644)
	fmt.Println("build number", bn)
}

func build(name, goos, goarch, target string, cgo int) {
	t := time.Now()
	fmt.Println("building", name)
	if runtime.GOOS == "windows" {
		c := fmt.Sprintf(`$env:GOOS="%s"; $env:GOARCH="%s"; $env:CGO_ENABLED=%d; $env:CC="%s cc -g0 -static -target %s"; go build -o %s -ldflags '-w -s -extldflags "-static"'`,
			goos, goarch, cgo, winZigPath, target, outputDir+name)
		// fmt.Println(c)
		exec.Command("powershell", "/c", c).Run()
	} else {
		c := fmt.Sprintf(`GOOS=%s GOARCH=%s CGO_ENABLED=%d CC="zig cc -g0 -static -target %s" go build -o %s -ldflags '-w -s -extldflags "-static"'`,
			goos, goarch, cgo, target, outputDir+name)
		// fmt.Println(c)
		exec.Command("bash", "-c", c).Run()
	}
	fmt.Println("\ttime :", time.Since(t).Seconds(), "sec")
}

func copyFile(src, dest string) {
	input, err := os.ReadFile(src)
	if err != nil {
		fmt.Println(err)
		return
	}

	os.MkdirAll(path.Dir(dest), 0777)

	err = os.WriteFile(dest, input, 0644)
	if err != nil {
		fmt.Println("Error creating", dest)
		fmt.Println(err)
		return
	}
}
