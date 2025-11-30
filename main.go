package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lucky-peak/nmtscope/server/config"
	"github.com/lucky-peak/nmtscope/server/handler"
	"github.com/lucky-peak/nmtscope/server/nmt"

	"github.com/robfig/cron/v3"
)

func main() {
	parseFlags()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM) // 监听中断信号 (Ctrl+C) 和终止信号。

	c := newCronJob()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/nmt", handler.NMTHandler)
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.CONFIG.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}
	mux.Handle("/", SpaHandler())

	go func() {
		log.Printf("Go server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Could not listen on %s: %v", srv.Addr, err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	if c != nil {
		c.Stop().Done()
		log.Println("Cron job stopped gracefully.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting gracefully.")
}

func newCronJob() *cron.Cron {
	var c *cron.Cron
	if config.CONFIG.Interval > 0 {
		c = cron.New(cron.WithSeconds())
		_, err := c.AddFunc(fmt.Sprintf("*/%d * * * * *", config.CONFIG.Interval), func() {
			log.Printf("Collecting NMT metrics for pid %d", config.CONFIG.Pid)
			nmt.GenerateNMTReport(config.CONFIG.Jcmd, config.CONFIG.Pid, config.CONFIG.ReportDir)
		})
		if err != nil {
			log.Fatalf("Error adding cron job: %v", err)
		}
		c.Start()
	}
	return c
}

func parseFlags() {
	flag.StringVar(&config.CONFIG.ReportDir, "report_dir", config.CONFIG.ReportDir, "Directory to store NMT reports")
	flag.IntVar(&config.CONFIG.Port, "port", config.CONFIG.Port, "Port to listen on")
	flag.IntVar(&config.CONFIG.Interval, "interval", config.CONFIG.Interval, "Interval in seconds to collect NMT metrics")
	flag.IntVar(&config.CONFIG.Retention, "retention", config.CONFIG.Retention, "Retention time in minutes to keep NMT metrics")
	flag.StringVar(&config.CONFIG.Jcmd, "jcmd", config.CONFIG.Jcmd, "Path to jcmd binary")
	flag.IntVar(&config.CONFIG.Pid, "pid", config.CONFIG.Pid, "PID of the Java process to monitor")

	flag.Parse()
	log.Printf("config: %+v", config.CONFIG)

	if config.CONFIG.Pid <= 0 {
		log.Fatalf("Invalid PID: %d", config.CONFIG.Pid)
	}
}
