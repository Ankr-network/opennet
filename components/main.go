package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Ankr-network/opennet/components/app"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	sockPath = "/tmp/opennet.sock"
)

var (
	endpoints = flag.String("endpoints",
		"https://10.51.24.207:2379,https://10.51.24.209:2379,https://10.51.24.214:2379",
		"etcd endpoints")
	debug = flag.Bool("debug", false, "sets log level to debug")
)

func main() {

	flag.Parse()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	app.Endpoints = strings.Split(*endpoints, ",")
	if err := app.InitStore(); err != nil {
		log.Error().Msg(err.Error())
		return
	}

	http.Handle("/ip", http.HandlerFunc(app.IPHandler))

	lis, err := getUnixListener()
	if err != nil {
		log.Error().Msg(err.Error())
		return
	}
	defer func() {
		defer lis.Close()
	}()

	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		select {
		case <-stopChan:
			time.Sleep(200 * time.Millisecond)
			if err := os.Remove(sockPath); err != nil {
				log.Error().Msg(err.Error())
			}
			log.Info().Msg("service exited totally.")
			os.Exit(0)
		}
	}()
	log.Info().Msg(fmt.Sprintf("server start with unix path: %s", sockPath))
	log.Fatal().Msg(http.Serve(lis, nil).Error())

}

func getUnixListener() (*net.UnixListener, error) {
	addr, err := net.ResolveUnixAddr("unix", sockPath)
	if err != nil {
		panic("Cannot resolve unix addr: " + err.Error())
	}
	return net.ListenUnix("unix", addr)
}
