package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

func read(path string) map[string]any {
	f, err := os.Open(path)
	must(err)
	defer f.Close()
	d := json.NewDecoder(f)
	d.UseNumber()
	var v map[string]any
	must(d.Decode(&v))
	return v
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
func at(v any, keys ...string) any {
	for _, k := range keys {
		v = v.(map[string]any)[k]
	}
	return v
}
func same(label string, a, b any) {
	if !reflect.DeepEqual(a, b) {
		panic("different: " + label)
	}
}
func main() {
	if len(os.Args) != 3 {
		panic("provide two observation directories")
	}
	a, b := os.Args[1], os.Args[2]
	x, y := read(filepath.Join(a, "summary.json")), read(filepath.Join(b, "summary.json"))
	for _, k := range []string{"functions", "consumer_programs", "model_observed"} {
		same(k, x[k], y[k])
	}
	xr, yr := x["runs"].([]any), y["runs"].([]any)
	if len(xr) != 25 || len(yr) != 25 {
		panic("expected 25 function/mode/budget runs")
	}
	for i := range xr {
		for _, k := range []string{"package", "activity", "mode", "budget", "passed", "input_tuples", "fields_passed", "fields_total", "generated_sha256", "replay_equal", "input_only_rows"} {
			same(fmt.Sprintf("run %d %s", i, k), at(xr[i], k), at(yr[i], k))
		}
	}
	dirs, err := os.ReadDir(a)
	must(err)
	observations := 0
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		for _, name := range []string{"execution.json", "replay.json", "input-only.json"} {
			pa, pb := filepath.Join(a, d.Name(), name), filepath.Join(b, d.Name(), name)
			_, ea := os.Stat(pa)
			_, eb := os.Stat(pb)
			if os.IsNotExist(ea) && os.IsNotExist(eb) {
				continue
			}
			must(ea)
			must(eb)
			ra, rb := read(pa), read(pb)
			same(d.Name()+name+" decision", ra["decision"], rb["decision"])
			for _, key := range []string{"traces", "finite_passed", "finite_total", "model_calls", "projection_replayed", "runtime_replayed", "generated_sha256"} {
				same(d.Name()+name+key, at(ra, "result", "runtime", key), at(rb, "result", "runtime", key))
			}
			observations++
		}
	}
	if observations != 52 {
		panic(fmt.Sprintf("expected 52 raw observations, got %d", observations))
	}
	fmt.Println("PASS: 25 run summaries and 52 raw observations agree on delivered inputs, actual/expected values, finite counts, generated code and saved replay. Timings and platform binary hashes are not compared.")
}
