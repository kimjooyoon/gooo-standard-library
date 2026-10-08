package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	compiler := flag.String("compiler", "gooo", "Gooo compiler executable")
	out := flag.String("out", "", "new observation directory")
	model := flag.String("model", "", "optional external graph model JSON")
	flag.Parse()
	if *out == "" {
		return fmt.Errorf("provide --out for a new directory")
	}
	root, err := filepath.Abs(*out)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(root), 0755); err != nil {
		return err
	}
	if err = os.Mkdir(root, 0755); err != nil {
		return err
	}
	if *model != "" {
		*model, err = filepath.Abs(*model)
		if err != nil {
			return err
		}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()
	if _, err = execute(ctx, *compiler, filepath.Join(root, "compiler-build.json"), "version", "--build", "--json"); err != nil {
		return err
	}
	var library []fixture
	if err = readJSON("fixtures/library.json", &library); err != nil {
		return err
	}
	if len(library) != 17 {
		return fmt.Errorf("expected the frozen 17-function roster")
	}
	var reports []report
	for _, f := range library {
		r, err := observe(ctx, *compiler, root, f, "library", 0, "")
		if err != nil {
			return err
		}
		if r.Passed != r.Total {
			return fmt.Errorf("%s:%s incomplete: %d/%d", f.Package, f.Activity, r.Passed, r.Total)
		}
		reports = append(reports, r)
	}
	var preview fixture
	if err = readJSON("fixtures/preview.json", &preview); err != nil {
		return err
	}
	modes := []string{"fixed"}
	if *model != "" {
		modes = append(modes, "model")
	}
	for _, mode := range modes {
		for _, budget := range []int{1, 2, 4, 8} {
			selectedModel := ""
			if mode == "model" {
				selectedModel = *model
			}
			r, err := observe(ctx, *compiler, root, preview, mode, budget, selectedModel)
			if err != nil {
				return err
			}
			if budget == 8 && r.Passed != r.Total {
				return fmt.Errorf("full preview budget incomplete: %d/%d", r.Passed, r.Total)
			}
			reports = append(reports, r)
		}
	}
	return save(filepath.Join(root, "summary.json"), map[string]any{
		"schema": "gooo/standard-library-observation/v1", "functions": 17,
		"consumer_programs": 1, "model_observed": *model != "", "runs": reports,
		"scope": "Finite explicit library cases and one importing consumer at four budgets; saved replay repeats the same inputs. No model training or broad intent-completion probability.",
	})
}
