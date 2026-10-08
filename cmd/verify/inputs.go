package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

func disjointPreview(f fixture) error {
	source, err := os.ReadFile("examples/preview.gooo")
	if err != nil {
		return err
	}
	matches := regexp.MustCompile(`value_case\s+("(?:[^"\\]|\\.)*")`).FindAllSubmatch(source, -1)
	if len(matches) != 5 {
		return fmt.Errorf("expected five frozen selection cases")
	}
	seen := map[string]bool{}
	for _, match := range matches {
		value, err := strconv.Unquote(string(match[1]))
		if err != nil {
			return err
		}
		decoded, err := decode(json.RawMessage(value))
		if err != nil {
			return err
		}
		canonical, _ := json.Marshal(decoded)
		seen[string(canonical)] = true
	}
	for _, row := range f.Cases {
		encoded, _ := json.Marshal(row.Inputs)
		decoded, err := decode(encoded)
		if err != nil {
			return err
		}
		canonical, _ := json.Marshal(decoded)
		if seen[string(canonical)] {
			return fmt.Errorf("preview evaluation repeats a selection or earlier evaluation input")
		}
		seen[string(canonical)] = true
	}
	return nil
}

func observeInputs(ctx context.Context, compiler, dir, generated string) (int, error) {
	var request struct {
		Schema string                       `json:"schema"`
		Inputs []map[string]json.RawMessage `json:"inputs"`
	}
	if err := readJSON("examples/preview-inputs.json", &request); err != nil {
		return 0, err
	}
	if request.Schema != "gooo/body-composition-inputs/v1" || len(request.Inputs) != 3 {
		return 0, fmt.Errorf("expected three usage inputs")
	}
	path := filepath.Join(dir, "inputs.json")
	if err := save(path, request); err != nil {
		return 0, err
	}
	raw, err := execute(ctx, compiler, filepath.Join(dir, "input-only.json"), "package", "replay", "--json", "--receipt", filepath.Join(dir, "execution.json"), "--inputs", path, filepath.Join(dir, "gooo.workspace.json"))
	if err != nil {
		return 0, err
	}
	var e envelope
	if err = json.Unmarshal(raw, &e); err != nil {
		return 0, err
	}
	r := e.Result.Runtime
	if e.Decision != "OBSERVED" || e.Error != "" || e.ReplayedFrom == "" || e.Result.Replay == nil || e.Result.Replay.Calls == nil || *e.Result.Replay.Calls != 0 || r.Calls == nil || *r.Calls != 0 || r.Passed == nil || *r.Passed != 0 || r.Total == nil || *r.Total != 0 || !r.Projection || !r.Replayed || e.Result.Composition.SHA != generated || len(r.Traces) != 3 {
		return 0, fmt.Errorf("input-only replay metadata differs")
	}
	for i, t := range r.Traces {
		if t.Index != i || len(t.Deliveries) != 1 || len(t.Deliveries[0].Expected) != 0 || t.Deliveries[0].Passed != nil || len(t.Deliveries[0].Actual) == 0 {
			return 0, fmt.Errorf("input-only execution acquired expectations or lost values")
		}
		d := t.Deliveries[0]
		if len(d.Inputs) != 2 {
			return 0, fmt.Errorf("usage input arity differs")
		}
		for j, input := range d.Inputs {
			if input.Port != fmt.Sprintf("input%d", j) || !equalJSON(input.Value, request.Inputs[i][fmt.Sprintf("examples/preview:Plan.input%d", j)]) {
				return 0, fmt.Errorf("usage argument differs")
			}
		}
	}
	return 3, nil
}
