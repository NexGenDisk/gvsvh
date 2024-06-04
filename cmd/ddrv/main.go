package main

import (
	"fmt"
	"log"
	"runtime"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/joho/godotenv"

	"github.com/NexGenDisk/gvsvh/internal/config"
	"github.com/NexGenDisk/gvsvh/internal/dataprovider"
	"github.com/NexGenDisk/gvsvh/internal/filesystem"
	"github.com/NexGenDisk/gvsvh/internal/ftp"
	"github.com/NexGenDisk/gvsvh/internal/http"
	"github.com/NexGenDisk/gvsvh/internal/webdav"
	"github.com/NexGenDisk/gvsvh/pkg/ddrv"
)

func main() {
	// Set the maximum number of operating system threads to use.
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Load env file.
	_ = godotenv.Load()

	// Parse command line arguments into config
	kong.Parse(config.New(), kong.Vars{
		"version": fmt.Sprintf("ddrv %s", version),
	})

	// Make sure chunkSize is below 25MB
	if config.ChunkSize() > 25*1024*1024 || config.ChunkSize() < 0 {
		log.Fatalf("ddrv: invalid chunkSize %d", config.ChunkSize())
	}

	// Create a ddrv manager
	mgr, err := ddrv.NewManager(config.ChunkSize(), strings.Split(config.Webhooks(), ","))
	if err != nil {
		log.Fatalf("ddrv: failed to open ddrv mgr :%v", err)
	}

	// Create FS object
	fs := filesystem.New(mgr)

	// New data provider
	dataprovider.New()

	errCh := make(chan error)

	if config.FTPAddr() != "" {
		go func() {
			// Create and start ftp server
			ftpServer := ftp.New(fs)
			log.Printf("ddrv: starting FTP server on : %s", config.FTPAddr())
			errCh <- ftpServer.ListenAndServe()
		}()
	}
	if config.HTTPAddr() != "" {
		go func() {
			httpServer := http.New(mgr)
			log.Printf("ddrv: starting HTTP server on : %s", config.HTTPAddr())
			errCh <- httpServer.Listen(config.HTTPAddr())
		}()
	}

	if config.WDAddr() != "" {
		go func() {
			webdavServer := webdav.New(fs)
			log.Printf("ddrv: starting WEBDAV server on : %s", config.WDAddr())
			errCh <- webdavServer.ListenAndServe()
		}()
	}

	log.Fatalf("ddrv: ddrv error %v", <-errCh)
}
