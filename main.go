package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/csrf"
	"github.com/kovetskiy/lorg"
	"github.com/kozhevnikova/channellogger"
)

var channelLogger *channellogger.ChannelData

func main() {

	lorg.SetFormat(
		lorg.NewFormat(`${level:[%s]:left}${file}::${line} %s`),
	)

	config, err := ParseConfig()
	if err != nil {
		lorg.Error(err)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error())
		return
	}

	channelLogger = &channellogger.ChannelData{
		Token:     config.Channellogger.Token,
		ChannelID: config.Channellogger.ChannelID,
	}

	server, err := NewServer(config)
	if err != nil {
		lorg.Error(err)
		channellogger.SendLogInfoToChannel(channelLogger, "APP::"+err.Error())
		return
	}

	router := NewRouter(server)

	CSRF := csrf.Protect([]byte(config.Csrf.Token),
		csrf.Secure(false))

	srv := &http.Server{
		Addr:         config.Server.Address,
		Handler:      CSRF(router),
		WriteTimeout: config.Server.WriteTimeout * time.Second,
		ReadTimeout:  config.Server.ReadTimeout * time.Second,
		IdleTimeout:  config.Server.IdleTimeout * time.Second,
	}

	lorg.Info("OK")

	err = srv.ListenAndServe()
	if err != nil {
		lorg.Error(err)
		channellogger.SendLogInfoToChannel(channelLogger,
			"APP::EXIT::"+err.Error())
		os.Exit(1)
	}
}
