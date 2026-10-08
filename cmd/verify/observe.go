package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
)

type report struct {
	Package      string            `json:"package"`
	Activity     string            `json:"activity"`
	Mode         string            `json:"mode"`
	Budget       int               `json:"budget"`
	Passed       int               `json:"passed"`
	Total        int               `json:"input_tuples"`
	FieldsPassed int               `json:"fields_passed"`
	FieldsTotal  int               `json:"fields_total"`
	GeneratedSHA string            `json:"generated_sha256"`
	ReplayEqual  bool              `json:"replay_equal"`
	Assembly     []json.RawMessage `json:"assembly"`
	Model        json.RawMessage   `json:"model"`
	InputOnly    int               `json:"input_only_rows,omitempty"`
}

type envelope struct {
	Schema, Decision, Error string
	ReplayedFrom            string `json:"replayed_from_sha256"`
	Result                  struct {
		Program struct {
			Entry struct {
				Package  string `json:"package_path"`
				Activity string `json:"activity"`
			} `json:"entry"`
		} `json:"program"`
		Replay *struct {
			Calls *int `json:"model_calls"`
		} `json:"replay"`
		Composition struct {
			SHA   string          `json:"generated_sha256"`
			Model json.RawMessage `json:"model"`
			Steps []struct {
				Generation struct {
					Report struct {
						Assembly json.RawMessage `json:"record_assembly"`
						ID       string          `json:"activity_id"`
					} `json:"report"`
				} `json:"generation"`
			} `json:"steps"`
		} `json:"composition"`
		Runtime struct {
			Calls      *int `json:"model_calls"`
			Passed     *int `json:"finite_passed"`
			Total      *int `json:"finite_total"`
			Projection bool `json:"projection_replayed"`
			Replayed   bool `json:"runtime_replayed"`
			Traces     []struct {
				Index      int `json:"case_index"`
				Deliveries []struct {
					Actual, Expected json.RawMessage
					Passed           *bool
					ID               string          `json:"activity_id"`
					Input            json.RawMessage `json:"input"`
					Inputs           []struct {
						Port  string          `json:"port"`
						Value json.RawMessage `json:"value"`
					} `json:"inputs"`
				} `json:"deliveries"`
			} `json:"traces"`
		} `json:"runtime"`
	} `json:"result"`
}

func execute(ctx context.Context, compiler, path string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, compiler, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if writeErr := os.WriteFile(path, stdout.Bytes(), 0644); writeErr != nil {
		return nil, writeErr
	}
	if stderr.Len() > 0 {
		if writeErr := os.WriteFile(path+".stderr", stderr.Bytes(), 0644); writeErr != nil {
			return nil, writeErr
		}
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %s", path, err, stderr.String())
	}
	return stdout.Bytes(), nil
}

func decode(raw json.RawMessage) (any, error) {
	var value any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	err := d.Decode(&value)
	return value, err
}

func inspect(raw []byte, f fixture) (envelope, int, int, int, []any, error) {
	var e envelope
	if err := json.Unmarshal(raw, &e); err != nil {
		return e, 0, 0, 0, nil, err
	}
	r := e.Result.Runtime
	if e.Result.Program.Entry.Package != f.Package || e.Result.Program.Entry.Activity != f.Activity || len(e.Result.Composition.Steps) != 1 {
		return e, 0, 0, 0, nil, fmt.Errorf("entry or constructed activity roster differs")
	}
	if e.Schema != "gooo/workspace-body-execution-receipt/v1" || e.Error != "" || r.Calls == nil || *r.Calls != 0 || r.Passed == nil || r.Total == nil || !r.Projection || !r.Replayed || len(r.Traces) != len(f.Cases) || e.Result.Composition.SHA == "" {
		return e, 0, 0, 0, nil, fmt.Errorf("incomplete execution evidence for %s:%s", f.Package, f.Activity)
	}
	passed, fp, ft := 0, 0, 0
	var actuals []any
	for i, t := range r.Traces {
		if t.Index != i || len(t.Deliveries) != 1 {
			return e, 0, 0, 0, nil, fmt.Errorf("unexpected case or delivery roster")
		}
		d := t.Deliveries[0]
		if d.ID == "" || d.ID != e.Result.Composition.Steps[0].Generation.Report.ID {
			return e, 0, 0, 0, nil, fmt.Errorf("delivery identity differs")
		}
		if len(f.Cases[i].Inputs) == 1 {
			if !equalJSON(d.Input, f.Cases[i].Inputs[0]) {
				return e, 0, 0, 0, nil, fmt.Errorf("delivered input differs")
			}
		} else {
			if len(d.Inputs) != len(f.Cases[i].Inputs) {
				return e, 0, 0, 0, nil, fmt.Errorf("input arity differs")
			}
			for j, input := range d.Inputs {
				if input.Port != fmt.Sprintf("input%d", j) || !equalJSON(input.Value, f.Cases[i].Inputs[j]) {
					return e, 0, 0, 0, nil, fmt.Errorf("delivered argument differs")
				}
			}
		}
		actual, err := decode(d.Actual)
		if err != nil {
			return e, 0, 0, 0, nil, err
		}
		expected, err := decode(f.Cases[i].Expected)
		if err != nil {
			return e, 0, 0, 0, nil, err
		}
		recorded, err := decode(d.Expected)
		if err != nil {
			return e, 0, 0, 0, nil, err
		}
		matched := reflect.DeepEqual(actual, expected)
		if !reflect.DeepEqual(recorded, expected) || d.Passed == nil || *d.Passed != matched {
			return e, 0, 0, 0, nil, fmt.Errorf("fixture or passed flag mismatch")
		}
		if matched {
			passed++
		}
		if fields, ok := expected.(map[string]any); ok {
			values, ok := actual.(map[string]any)
			if !ok {
				return e, 0, 0, 0, nil, fmt.Errorf("record result required")
			}
			for key, want := range fields {
				ft++
				if value, present := values[key]; present && reflect.DeepEqual(value, want) {
					fp++
				}
			}
		}
		actuals = append(actuals, actual)
	}
	if passed != *r.Passed || len(f.Cases) != *r.Total {
		return e, 0, 0, 0, nil, fmt.Errorf("finite count mismatch")
	}
	wantDecision := "PROGRESS"
	if passed == len(f.Cases) {
		wantDecision = "PASS"
	}
	if e.Decision != wantDecision {
		return e, 0, 0, 0, nil, fmt.Errorf("execution decision differs from finite observations")
	}
	return e, passed, fp, ft, actuals, nil
}

func observe(ctx context.Context, compiler, root string, f fixture, mode string, budget int, model string) (report, error) {
	r := report{Package: f.Package, Activity: f.Activity, Mode: mode, Budget: budget, Total: len(f.Cases)}
	dir, err := materialize(root, f, mode, budget)
	if err != nil {
		return r, err
	}
	manifest, cases := filepath.Join(dir, "gooo.workspace.json"), filepath.Join(dir, "cases.json")
	args := []string{"package", "execute", "--json", "--cases", cases}
	if model != "" {
		args = append(args, "--assembly-model", model)
	}
	args = append(args, manifest)
	raw, err := execute(ctx, compiler, filepath.Join(dir, "execution.json"), args...)
	if err != nil {
		return r, err
	}
	built, passed, fp, ft, actuals, err := inspect(raw, f)
	if err != nil {
		return r, err
	}
	r.Passed, r.FieldsPassed, r.FieldsTotal, r.GeneratedSHA = passed, fp, ft, built.Result.Composition.SHA
	r.Model = built.Result.Composition.Model
	if built.ReplayedFrom != "" || built.Result.Replay != nil {
		return r, fmt.Errorf("execute unexpectedly returned saved-replay metadata")
	}
	for _, step := range built.Result.Composition.Steps {
		if a := step.Generation.Report.Assembly; len(a) > 0 && string(a) != "null" {
			r.Assembly = append(r.Assembly, a)
		}
	}
	if err = checkModelUse(r, model != ""); err != nil {
		return r, err
	}
	raw, err = execute(ctx, compiler, filepath.Join(dir, "replay.json"), "package", "replay", "--json", "--receipt", filepath.Join(dir, "execution.json"), "--cases", cases, manifest)
	if err != nil {
		return r, err
	}
	saved, _, _, _, replayed, err := inspect(raw, f)
	if err != nil {
		return r, err
	}
	if saved.Result.Replay == nil || saved.Result.Replay.Calls == nil || *saved.Result.Replay.Calls != 0 || saved.ReplayedFrom == "" || saved.Result.Composition.SHA != r.GeneratedSHA || !reflect.DeepEqual(actuals, replayed) {
		return r, fmt.Errorf("saved replay differs")
	}
	r.ReplayEqual = true
	if budget == 8 {
		r.InputOnly, err = observeInputs(ctx, compiler, dir, r.GeneratedSHA)
		if err != nil {
			return r, err
		}
	}
	fmt.Printf("%s:%s %s budget=%d cases=%d/%d fields=%d/%d replay=equal\n", f.Package, f.Activity, mode, budget, r.Passed, r.Total, fp, ft)
	return r, nil
}

func equalJSON(left, right json.RawMessage) bool {
	l, le := decode(left)
	r, re := decode(right)
	return le == nil && re == nil && reflect.DeepEqual(l, r)
}

func checkModelUse(r report, enabled bool) error {
	if r.Budget == 0 && !enabled && len(r.Assembly) == 0 && len(r.Model) == 0 {
		return nil
	}
	var m struct {
		Loaded *bool `json:"loaded"`
	}
	if err := json.Unmarshal(r.Model, &m); err != nil {
		return err
	}
	if m.Loaded == nil || *m.Loaded != enabled {
		return fmt.Errorf("model load observation differs")
	}
	want := 0
	if r.Budget > 0 {
		want = 1
	}
	if len(r.Assembly) != want {
		return fmt.Errorf("assembly activity count differs")
	}
	for _, raw := range r.Assembly {
		var a struct {
			Calls *int `json:"model_calls"`
			Total int  `json:"total"`
		}
		if err := json.Unmarshal(raw, &a); err != nil {
			return err
		}
		calls := 0
		if enabled {
			calls = 1
		}
		if a.Calls == nil || *a.Calls != calls || a.Total != 5 {
			return fmt.Errorf("construction call or selection roster differs")
		}
	}
	return nil
}
