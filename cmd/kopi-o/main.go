// Command kopi-o is a Caffeine-style menu-bar sleep holder.
//
// It restores the persisted preset from config on launch (Off = idle),
// holds idle sleep via a /usr/bin/caffeinate child process, and shows a
// countdown in the menu bar. Flags:
//
//	-config PATH   use PATH as the config file (else XDG/defaults)
//	-foreground    run attached in the foreground with log output
package main

import (
	"flag"
	"log"
	"os"

	"kopio/internal/config"
	"kopio/internal/sleep"
	"kopio/internal/tray"
)

func main() {
	configPath := flag.String("config", "", "config file path (overrides KOPI_O_CONFIG)")
	foreground := flag.Bool("foreground", false, "run attached in the foreground")
	flag.Parse()

	if *configPath != "" {
		if err := os.Setenv("KOPI_O_CONFIG", *configPath); err != nil {
			log.Fatalf("kopi-o: set config override: %v", err)
		}
	}
	if !*foreground {
		log.SetOutput(os.Stderr)
	}
	log.Printf("kopi-o: starting (config %s)", config.Path())

	sup := sleep.New(tray.Refresh)
	tray.Run(tray.Deps{Sup: sup})
}
