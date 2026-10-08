package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type options struct {
	compiler, title, out, model, receipt string
	budget                               int64
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parse(args []string, stderr io.Writer) (options, error) {
	var o options
	f := flag.NewFlagSet("preview", flag.ContinueOnError)
	f.SetOutput(stderr)
	f.StringVar(&o.compiler, "compiler", "gooo", "Gooo compiler executable")
	f.StringVar(&o.title, "title", "", "title passed unchanged to the Gooo program")
	f.Int64Var(&o.budget, "budget", 80, "byte budget passed unchanged to the Gooo program")
	f.StringVar(&o.out, "out", "", "new directory for inputs and the execution receipt")
	f.StringVar(&o.model, "model", "", "optional local model for fresh construction")
	f.StringVar(&o.receipt, "receipt", "", "saved execution to reuse without new inference")
	if err := f.Parse(args); err != nil {
		return o, err
	}
	if o.out == "" || f.NArg() != 0 {
		return o, fmt.Errorf("run from the repository root with --out pointing to a new directory")
	}
	if o.model != "" && o.receipt != "" {
		return o, fmt.Errorf("--receipt reuses saved choices; omit --model when replaying")
	}
	return o, nil
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	o, err := parse(args, stderr)
	if err != nil {
		return err
	}
	workspace, err := filepath.Abs("gooo.workspace.json")
	if err != nil {
		return err
	}
	if _, err = os.Stat(workspace); err != nil {
		return fmt.Errorf("run from the gooo-standard-library repository root: %w", err)
	}
	if err = os.MkdirAll(filepath.Dir(o.out), 0755); err != nil {
		return err
	}
	if err = os.Mkdir(o.out, 0755); err != nil {
		return fmt.Errorf("create new output directory: %w", err)
	}
	input, err := json.MarshalIndent(inputRequest(o), "", "  ")
	if err != nil {
		return err
	}
	inputPath := filepath.Join(o.out, "inputs.json")
	if err = os.WriteFile(inputPath, append(input, '\n'), 0644); err != nil {
		return err
	}
	raw, err := invoke(ctx, o, inputPath, workspace)
	if err != nil {
		return err
	}
	actual, err := actualValue(raw, o)
	if err != nil {
		return err
	}
	fmt.Fprintln(stderr, "Gooo execution receipt:", filepath.Join(o.out, "execution.json"))
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(actual)
}

func inputRequest(o options) any {
	return map[string]any{
		"schema": "gooo/body-composition-inputs/v1",
		"inputs": []map[string]any{{
			"examples/preview:Plan.input0": o.title,
			"examples/preview:Plan.input1": o.budget,
		}},
	}
}

func invoke(ctx context.Context, o options, input, workspace string) ([]byte, error) {
	args := []string{"package", "execute", "--json", "--inputs", input}
	if o.receipt != "" {
		args[1] = "replay"
		args = append(args, "--receipt", o.receipt)
	}
	if o.model != "" {
		args = append(args, "--assembly-model", o.model)
	}
	cmd := exec.CommandContext(ctx, o.compiler, append(args, workspace)...)
	path := filepath.Join(o.out, "compiler.stderr")
	stderr, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer stderr.Close()
	cmd.Stderr = stderr
	raw, runErr := cmd.Output()
	if err := os.WriteFile(filepath.Join(o.out, "execution.json"), raw, 0644); err != nil {
		return nil, err
	}
	if runErr != nil {
		return nil, fmt.Errorf("Gooo execution: %w; diagnostics: %s", runErr, path)
	}
	return raw, nil
}
