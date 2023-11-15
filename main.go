package main

import (
	"FGateWay/dao"
	"FGateWay/golang_common/lib"
	"FGateWay/http_proxy_router"
	"FGateWay/router"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

var (
	endpoint = flag.String("endpoint", "", "input endpoint dashboard or server")
	config   = flag.String("config", "", "input config file like ./conf/dev/")
)

func main() {
	flag.Parse()
	if *endpoint == "" {
		flag.Usage()
		os.Exit(1)
	}
	if *config == "" {
		flag.Usage()
		os.Exit(1)
	}

	if *endpoint == "dashboard" {
		lib.InitModule(*config)
		defer lib.Destroy()
		router.HttpServerRun()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		router.HttpServerStop()
	} else {
		lib.InitModule(*config)
		defer lib.Destroy()
		dao.ServiceManagerHandler.LoadOnce()
		//dao.AppManagerHandler.LoadOnce()

		go func() {
			http_proxy_router.HttpServerRun()
		}()
		go func() {
			http_proxy_router.HttpsServerRun()
		}()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		http_proxy_router.HttpServerStop()
		http_proxy_router.HttpsServerStop()

	}
}
