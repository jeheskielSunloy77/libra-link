package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/api"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/app"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/config"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/session"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/storage/repo"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/storage/sqlite"
	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/storage/sqlite/sqlcdb"
	syncer "github.com/jeheskielSunloy77/libra-link/apps/tui/internal/sync"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fatalf("load config: %v", err)
	}

	db, err := sqlite.Open(cfg.DBPath)
	if err != nil {
		fatalf("open sqlite: %v", err)
	}
	defer db.Close()

	queries := sqlcdb.New(db)
	repository := repo.New(queries)

	apiClient, err := api.NewClient(cfg.APIBaseURL, cfg.HTTPTimeout)
	if err != nil {
		fatalf("create API client: %v", err)
	}

	sessionStore := session.NewStore(cfg.SessionPath)
	worker := syncer.NewWorker(repository, apiClient, cfg.SyncBatchSize)
	model := app.New(cfg, apiClient, repository, sessionStore, worker)

	program := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := program.Run(); err != nil {
		fatalf("run tui: %v", err)
	}
}

func fatalf(pattern string, args ...any) {
	fmt.Fprintf(os.Stderr, pattern+"\n", args...)
	os.Exit(1)
}
