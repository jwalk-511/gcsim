package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	selfupdate "github.com/creativeprojects/go-selfupdate"
	"github.com/genshinsim/gcsim/pkg/servermode"
)

var (
	shareKey string
	version  string
)

type opts struct {
	host        string
	port        string
	shareKey    string
	timeout     int
	workers     int
	update      bool
	showVersion bool
}

func main() {
	if shareKey == "" {
		shareKey = os.Getenv("GCSIM_SHARE_KEY")
	}

	var opt opts
	flag.StringVar(&opt.host, "host", "localhost", "host to listen to (default: localhost)")
	flag.StringVar(&opt.port, "port", "54321", "port to listen on (default: 54321)")
	flag.StringVar(&opt.shareKey, "sharekey", "", "share key to use (default: build flag OR GCSIM_SHARE_KEY env variable if not available)")
	flag.IntVar(&opt.timeout, "timeout", 5*60, "how long to run each sim for in seconds before timing out (default: 300s)")
	flag.IntVar(&opt.workers, "workers", 10, "how many workers to use (default: 10)")
	flag.BoolVar(&opt.update, "update", false, "run autoupdater (default: false)")
	flag.BoolVar(&opt.showVersion, "version", false, "show currrent version")
	flag.Parse()

	if opt.showVersion {
		fmt.Println("Running version: ", version)
		return
	}

	if opt.update {
		err := update(version)
		if err != nil {
			fmt.Printf("Error running autoupdater: %v. Please update manually or run this executable with -update=false to skip autoupdate\n", err)
			fmt.Print("Press 'Enter' to exit...")
			bufio.NewReader(os.Stdin).ReadBytes('\n')
			os.Exit(1)
		}
	}

	if opt.shareKey != "" {
		shareKey = opt.shareKey
	}

	server, err := servermode.New(
		servermode.WithDefaults(),
		servermode.WithShareKey(shareKey),
		servermode.WithWorkers(opt.workers),
		servermode.WithTimeout(time.Duration(opt.timeout)*time.Second),
	)

	if err != nil {
		panic(err)
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf("%v:%v", opt.host, opt.port), server.Router))
}

func update(version string) error {
	src, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return fmt.Errorf("error creating GitHub source: %w", err)
	}
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:  src,
		Filters: []string{"server_.+"},
	})
	if err != nil {
		return fmt.Errorf("error creating updater: %w", err)
	}

	latest, found, err := updater.DetectLatest(context.Background(), selfupdate.ParseSlug("genshinsim/gcsim"))
	if err != nil {
		return fmt.Errorf("error occurred while detecting version: %w", err)
	}
	if !found {
		return fmt.Errorf("latest version for %s/%s could not be found from github repository", runtime.GOOS, runtime.GOARCH)
	}

	if latest.LessOrEqual(version) {
		log.Printf("Current version (%s) is the latest", version)
		return nil
	}

	log.Printf("Found latest version %v published at %v (%v), greater than current version %v\n", latest.Name, latest.PublishedAt, latest.AssetName, version)

	exe, err := os.Executable()
	if err != nil {
		return errors.New("could not locate executable path")
	}
	if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
		return fmt.Errorf("error occurred while updating binary: %w", err)
	}
	log.Printf("Successfully updated to version %s", latest.Version())
	return nil
}
