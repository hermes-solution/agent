package main

import (
	"context"
	"crypto/md5"
	"flag"
	"fmt"
	"go.etcd.io/etcd/clientv3"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"
)

/**
agent --use-file --file --use-etcd --etcd-address
 */
const (
	fluentdCommand = "fluentd"
	fluentdConfig  = "/hermes/fluentd.conf"
	originalConfig = "/hermes/forward.conf"
)

var (
	useFile      = false
	fileConfig   = ""
	useEtcd      = false
	etcdAddress  = ""
	etcdKey      = ""
	fileChecksum = ""
	cmd          *exec.Cmd
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

func start() {
	var err error
	if cmd != nil {
		err = KillProcess(cmd.Process.Pid)
		if err != nil {
			log.Println("kill old process get error", err)
		} else {
			log.Println("kill old process successful")
		}
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
}

func checksumAndStart() {
	checksum, err := hashMD5(fluentdConfig)
	if err != nil {
		log.Fatal(err)
	}
	if fileChecksum == "" || fileChecksum != checksum {
		start()
		fileChecksum = checksum
	}
}

func rewriteConfig(appendData []byte) error {
	data, err := ioutil.ReadFile(originalConfig)
	if err != nil {
		return err
	}

	if appendData != nil && len(appendData) > 0 {
		data = append(data, []byte("\n")...)
		data = append(data, appendData...)
	}
	log.Println("new configuration: ", string(data))
	err = ioutil.WriteFile(fluentdConfig, data, 0777)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	log.SetOutput(os.Stdout)
	flag.BoolVar(&useFile, "use-file", false, "using file config is passed via file argument")
	flag.StringVar(&fileConfig, "file", "", "absolute path of configuration file")
	flag.BoolVar(&useEtcd, "use-etcd", false, "using configuration is fetched from etcd")
	flag.StringVar(&etcdAddress, "etcd-address", "", "address of etcd")
	flag.StringVar(&etcdKey, "etcd-key", "/hermes/agent/config", "configuration key in etcd")
	flag.Parse()

	if useEtcd && etcdAddress != "" && etcdKey != "" {
		useFile = false
		fileConfig = ""
	}

	if useFile && fileConfig != "" {
		matchData, err := ioutil.ReadFile(fileConfig)
		if err != nil {
			log.Fatal(err)
		}

		err = rewriteConfig(matchData)
		if err != nil {
			log.Fatal(err)
		}

		go watchOnConfigFile()
	}
	if useEtcd && etcdAddress != "" && etcdKey != "" {
		matchData, err := readConfigFromEtcd()
		if err != nil {
			log.Fatal(err)
		}

		err = rewriteConfig(matchData)
		if err != nil {
			log.Fatal(err)
		}

		go watchOnEtcdConfigChange()
	}
	/**
	start fluentd and scheduler for reloading
	 */
	checksumAndStart()
	go scheduleReload()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, GetSingal()...)
	for {
		<-signalChannel
		signal.Stop(signalChannel)
		if cmd != nil {
			err := KillProcess(cmd.Process.Pid)
			if err != nil {
				log.Println(err)
			}
		}
		close(signalChannel)
		break
	}
}

func readConfigFromEtcd() ([]byte, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(etcdAddress, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = cli.Close()
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer func() {
		cancel()
	}()
	response, err := cli.Get(ctx, etcdKey)
	if err != nil {
		return nil, err
	}
	if response.Kvs != nil && len(response.Kvs) > 0 {
		for _, v := range response.Kvs {
			return v.Value, nil
		}
	}
	return nil, nil
}

func watchOnEtcdConfigChange() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(etcdAddress, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = cli.Close()
	}()

	rch := cli.Watch(context.Background(), etcdKey)
	for wresp := range rch {
		for _, ev := range wresp.Events {
			err = rewriteConfig(ev.Kv.Value)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func watchOnConfigFile() {
	originalSum, err := hashMD5(fileConfig)
	if err != nil {
		log.Fatal(err)
	}
	t := time.NewTicker(2 * time.Second)
	for {
		<-t.C
		newSum, err := hashMD5(fileConfig)
		if err != nil {
			log.Println("READ CONFIG FILE GET ERROR: ", err)
		}
		if newSum != originalSum {
			originalSum = newSum
			matchData, err := ioutil.ReadFile(fileConfig)
			if err != nil {
				log.Fatal(err)
			}
			err = rewriteConfig(matchData)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

func scheduleReload() {
	t := time.NewTicker(5000 * time.Millisecond)
	for {
		<-t.C
		checksumAndStart()
	}
}
