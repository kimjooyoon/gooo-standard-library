package main

import (
	"encoding/json"
	"testing"
)

func TestExactJSONComparison(t *testing.T) {
	if equalJSON(json.RawMessage(`9223372036854775807`), json.RawMessage(`9223372036854775806`)) {
		t.Fatal("integer precision lost")
	}
	if !equalJSON(json.RawMessage(`{"a":1,"b":false}`), json.RawMessage(`{"b":false,"a":1}`)) {
		t.Fatal("object order changed meaning")
	}
	if equalJSON(json.RawMessage(`null`), nil) {
		t.Fatal("missing JSON became null")
	}
}

func TestInspectBindsValuesAndMetadata(t *testing.T) {
	f := fixture{Package: "std/numbers", Activity: "Sign", Cases: []fixtureCase{{Inputs: []json.RawMessage{json.RawMessage(`1`)}, Expected: json.RawMessage(`1`)}}}
	base := `{"schema":"gooo/workspace-body-execution-receipt/v1","decision":"PASS","result":{"program":{"entry":{"package_path":"std/numbers","activity":"Sign"}},"composition":{"generated_sha256":"sha256:test","steps":[{"generation":{"report":{"activity_id":"sign"}}}]},"runtime":{"model_calls":0,"finite_passed":1,"finite_total":1,"projection_replayed":true,"runtime_replayed":true,"traces":[{"case_index":0,"deliveries":[{"activity_id":"sign","input":1,"actual":1,"expected":1,"passed":true}]}]}}}`
	if _, passed, _, _, _, err := inspect([]byte(base), f); err != nil || passed != 1 {
		t.Fatalf("valid observation: %d %v", passed, err)
	}
	mutations := map[string]func(map[string]any){
		"missing zero counter": func(r map[string]any) { delete(r["runtime"].(map[string]any), "model_calls") },
		"wrong entry":          func(r map[string]any) { r["program"].(map[string]any)["entry"].(map[string]any)["activity"] = "Other" },
		"wrong input":          func(r map[string]any) { delivery(r)["input"] = 2 },
		"wrong output":         func(r map[string]any) { delivery(r)["actual"] = 2 },
		"changed expectation":  func(r map[string]any) { delivery(r)["expected"] = 2 },
		"wrong identity":       func(r map[string]any) { delivery(r)["activity_id"] = "other" },
		"wrong count":          func(r map[string]any) { r["runtime"].(map[string]any)["finite_passed"] = 0 },
	}
	for name, change := range mutations {
		t.Run(name, func(t *testing.T) {
			var e map[string]any
			if err := json.Unmarshal([]byte(base), &e); err != nil {
				t.Fatal(err)
			}
			change(e["result"].(map[string]any))
			raw, _ := json.Marshal(e)
			if _, _, _, _, _, err := inspect(raw, f); err == nil {
				t.Fatal("invalid observation accepted")
			}
		})
	}
}

func delivery(r map[string]any) map[string]any {
	return r["runtime"].(map[string]any)["traces"].([]any)[0].(map[string]any)["deliveries"].([]any)[0].(map[string]any)
}
