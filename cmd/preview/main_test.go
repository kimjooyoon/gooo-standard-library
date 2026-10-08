package main

import (
	"encoding/json"
	"io"
	"math"
	"strings"
	"testing"
)

func TestArgumentsKeepGoooInputs(t *testing.T) {
	o, err := parse([]string{"--out", "new", "--title", "한글", "--budget", "-9223372036854775808"}, io.Discard)
	if err != nil || o.budget != math.MinInt64 || o.title != "한글" {
		t.Fatalf("arguments changed: %+v, %v", o, err)
	}
	b, err := json.Marshal(inputRequest(o))
	if err != nil || !json.Valid(b) {
		t.Fatalf("input request: %s, %v", b, err)
	}
	var input struct{ Inputs []map[string]json.RawMessage }
	if err = json.Unmarshal(b, &input); err != nil || len(input.Inputs) != 1 || string(input.Inputs[0]["examples/preview:Plan.input1"]) != "-9223372036854775808" {
		t.Fatalf("input precision changed: %s, %v", b, err)
	}
}

func TestInvalidArguments(t *testing.T) {
	for _, args := range [][]string{
		{}, {"--out", "new", "extra"}, {"--out", "new", "--model", "model", "--receipt", "receipt"},
	} {
		if _, err := parse(args, io.Discard); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestResultUsesActualGoooOutput(t *testing.T) {
	raw := []byte(`{"schema":"gooo/workspace-body-execution-receipt/v1","decision":"OBSERVED","result":{"program":{"entry":{"package_path":"examples/preview","activity":"Plan"}},"runtime":{"model_calls":0,"finite_passed":0,"finite_total":0,"traces":[{"case_index":0,"deliveries":[{"inputs":[{"port":"input0","value":"한글"},{"port":"input1","value":-9223372036854775808}],"actual":{"title":"from Gooo","bytes":9223372036854775807}}]}]}}}`)
	o := options{title: "한글", budget: math.MinInt64}
	actual, err := actualValue(raw, o)
	if err != nil || string(actual) != `{"title":"from Gooo","bytes":9223372036854775807}` {
		t.Fatalf("actual output changed: %s, %v", actual, err)
	}
	for _, change := range [][2]string{
		{`"model_calls":0`, `"unused":0`},
		{`"finite_total":0`, `"finite_total":1`},
		{`"case_index":0`, `"case_index":1`},
		{`"activity":"Plan"`, `"activity":"Other"`},
		{`"actual":{`, `"passed":true,"actual":{`},
	} {
		if _, err = actualValue([]byte(strings.Replace(string(raw), change[0], change[1], 1)), o); err == nil {
			t.Fatalf("accepted changed receipt: %v", change)
		}
	}
	o.receipt = "saved.json"
	if _, err = actualValue(raw, o); err == nil {
		t.Fatal("accepted missing saved-replay evidence")
	}
	o.receipt = ""
	o.budget++
	if _, err = actualValue(raw, o); err == nil {
		t.Fatal("accepted changed input")
	}
	if _, err = actualValue([]byte(`{"decision":"PASS"}`), o); err == nil {
		t.Fatal("accepted a result without actual execution")
	}
}
