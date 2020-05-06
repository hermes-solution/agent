package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

/**
agent --use-file --use-api --use-etcd --port --bind
 */
const (
	fluentdCommand = "fluentd"
	fluentdConfig  = "/fluentd/etc/fluentd.conf"
)

var (
	useFile         = false
	intervalRefresh = 1000
	useApi          = false
	useEtcd         = false
	apiPort         = 80
	fileChecksum    = ""
	cmd             *exec.Cmd
)

func hashMD5(file string) (string, error) {
	hasher := md5.New()
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = f.Close()
	}()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func checksumAndReload() {
	checksum, err := hashMD5(fluentdConfig)
	if err != nil {
		log.Fatal(err)
	}
	if fileChecksum == "" || fileChecksum != checksum {
		if cmd != nil {
			KillProcess(cmd.Process.Pid)
		}
		args := make([]string, 0)
		args = append(args, "-c")
		args = append(args, fluentdConfig)
		err = os.Setenv("FLUENTD_CONF", "fluentd.conf")
		cmd = exec.Command(fluentdCommand, args...)
		if err != nil {
			log.Fatal(err)
		}
		if cmd.Env == nil {
			cmd.Env = make([]string, 0)
		}
		cmd.Env = append(cmd.Env, "FLUENTD_CONF=fluentd.conf")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		fileChecksum = checksum
	}
}

func main() {
	log.SetOutput(os.Stdout)
	flag.BoolVar(&useFile, "use-file", true, "")
	flag.IntVar(&intervalRefresh, "interval-refresh", 1000, "")
	flag.BoolVar(&useApi, "use-api", false, "")
	flag.BoolVar(&useEtcd, "use-etcd", false, "")
	flag.IntVar(&apiPort, "api-port", 80, "")
	flag.Parse()
	if useFile {
		checksumAndReload()
		go func() {
			t := time.NewTicker(time.Duration(intervalRefresh) * time.Millisecond)
			for {
				<-t.C
				checksumAndReload()
			}
		}()
	}

	if useApi {
		http.HandleFunc("/config", func(writer http.ResponseWriter, request *http.Request) {
			
		})
		go func() {
			err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", apiPort), nil)
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	if useEtcd {
		go func() {

		}()
	}

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, GetSingal()...)
	for {
		<-signalChannel
		signal.Stop(signalChannel)
		if cmd != nil {
			KillProcess(cmd.Process.Pid)
		}
		break
	}
}
