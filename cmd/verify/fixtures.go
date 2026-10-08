package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type fixture struct {
	Package  string        `json:"package"`
	Activity string        `json:"activity"`
	Cases    []fixtureCase `json:"cases"`
}
type fixtureCase struct {
	Inputs   []json.RawMessage `json:"inputs"`
	Expected json.RawMessage   `json:"expected"`
}

func readJSON(path string, dest any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, dest)
}

func save(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0644)
}

func materialize(root string, f fixture, mode string, budget int) (string, error) {
	dir := filepath.Join(root, strings.ReplaceAll(f.Package, "/", "-")+"-"+f.Activity+"-"+mode+fmt.Sprint(budget))
	if err := os.Mkdir(dir, 0755); err != nil {
		return "", err
	}
	for _, path := range []string{"library/numbers.gooo", "library/logic.gooo", "library/text.gooo", "examples/preview.gooo"} {
		raw, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		if path == "examples/preview.gooo" && budget > 0 {
			if strings.Count(string(raw), `attempts "8"`) != 1 {
				return "", fmt.Errorf("unexpected preview budget declaration")
			}
			raw = []byte(strings.Replace(string(raw), `attempts "8"`, fmt.Sprintf(`attempts "%d"`, budget), 1))
		}
		target := filepath.Join(dir, path)
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", err
		}
		if err = os.WriteFile(target, raw, 0644); err != nil {
			return "", err
		}
	}
	var manifest map[string]json.RawMessage
	if err := readJSON("gooo.workspace.json", &manifest); err != nil {
		return "", err
	}
	entry, _ := json.Marshal(map[string]string{"package_path": f.Package, "activity": f.Activity})
	manifest["entry"] = entry
	if err := save(filepath.Join(dir, "gooo.workspace.json"), manifest); err != nil {
		return "", err
	}
	var cases []any
	for _, c := range f.Cases {
		if len(c.Inputs) == 0 || len(c.Expected) == 0 {
			return "", fmt.Errorf("fixture has missing input or expectation")
		}
		values := map[string]json.RawMessage{}
		key := f.Package + ":" + f.Activity
		for i, v := range c.Inputs {
			name := key
			if len(c.Inputs) > 1 {
				name = fmt.Sprintf("%s.input%d", key, i)
			}
			values[name] = v
		}
		cases = append(cases, map[string]any{"inputs": values, "expected": map[string]json.RawMessage{key: c.Expected}})
	}
	if err := save(filepath.Join(dir, "cases.json"), map[string]any{"schema": "gooo/body-composition-cases/v1", "cases": cases}); err != nil {
		return "", err
	}
	return dir, nil
}
