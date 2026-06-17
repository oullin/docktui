package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	dockerd "github.com/oullin/docktui/internal/docker"
	"github.com/oullin/docktui/internal/project"
	"github.com/oullin/docktui/internal/tui"
)

var version = "dev"

func main() {
	var (
		configPath = flag.String("config", "", "path to a docktui project config file")
		showVer    = flag.Bool("version", false, "print version and exit")
	)

	flag.Parse()

	if *showVer {
		fmt.Println(version)

		return
	}

	ctx := context.Background()
	client, err := dockerd.NewSDKClient()

	if err != nil {
		fmt.Fprintf(os.Stderr, "docktui: %v\n", err)
		os.Exit(1)
	}

	config, err := project.LoadConfig(*configPath)

	if err != nil {
		fmt.Fprintf(os.Stderr, "docktui: %v\n", err)
		os.Exit(1)
	}

	program := tea.NewProgram(tui.New(ctx, client, config), tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "docktui: %v\n", err)
		os.Exit(1)
	}
}
