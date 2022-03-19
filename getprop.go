package androidutils

import (
	"errors"
	"log"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	shellquote "github.com/kballard/go-shellquote"
)

const defaultShellTimeout = 10 * time.Second

var (
	isAndroid    = true
	deviceSerial = ""
)

func Initz(serial string, is bool) {
	isAndroid = is
	deviceSerial = serial
}

// run shell with default timeout
func runShell(args ...string) (out string, err error) {
	prefix := []string{}
	if !isAndroid {
		if deviceSerial != "" {
			prefix = []string{"adb", "-s", deviceSerial, "shell"}
		} else {
			prefix = []string{"adb", "shell"}
		}
	}
	sh := append(prefix, args...)
	cmd := exec.Command("sh", "-c", shellquote.Join(sh...))
	timer := time.AfterFunc(defaultShellTimeout, func() {
		cmd.Process.Kill()
	})
	defer timer.Stop()
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Properties extract info from $ adb shell getprop
func Properties() (props map[string]string, err error) {
	propOutput, err := runShell("getprop")
	if err != nil {
		return nil, err
	}
	re := regexp.MustCompile(`\[(.*?)\]:\s*\[(.*?)\]`)
	matches := re.FindAllStringSubmatch(propOutput, -1)
	props = make(map[string]string)
	for _, m := range matches {
		var key = m[1]
		var val = m[2]
		props[key] = val
	}
	return
}

var (
	propOnce   sync.Once
	properties map[string]string

	ErrGetprop = errors.New("error call getprop")
)

// Return property by name from cache
func CachedProperty(name string) string {
	propOnce.Do(func() {
		var err error
		properties, err = Properties()
		if err != nil {
			log.Println("getgrop", err)
		}
	})
	value := properties[name]
	if value == "" {
		value = Property(name)
	}
	return value
}

// Return property by name
func Property(name string) string {
	propOutput, err := runShell("getprop", name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(propOutput)
}
