package main

import (
	"encoding/json"
	"fmt"
)

func actualValue(raw []byte, o options) (json.RawMessage, error) {
	var e struct {
		Schema, Decision, Error string
		ReplayedFrom            string `json:"replayed_from_sha256"`
		Result                  struct {
			Replay *struct {
				Calls *int `json:"model_calls"`
			}
			Program struct {
				Entry struct {
					Package  string `json:"package_path"`
					Activity string
				}
			}
			Runtime struct {
				Calls  *int `json:"model_calls"`
				Passed *int `json:"finite_passed"`
				Total  *int `json:"finite_total"`
				Traces []struct {
					Index      int `json:"case_index"`
					Deliveries []struct {
						Actual, Expected json.RawMessage
						Passed           *bool
						Inputs           []struct {
							Port  string
							Value json.RawMessage
						}
					}
				}
			}
		}
	}
	if err := json.Unmarshal(raw, &e); err != nil {
		return nil, fmt.Errorf("read Gooo execution receipt: %w", err)
	}
	r := e.Result.Runtime
	if o.receipt != "" && (e.ReplayedFrom == "" || e.Result.Replay == nil || e.Result.Replay.Calls == nil || *e.Result.Replay.Calls != 0) {
		return nil, fmt.Errorf("expected a saved program replay with zero new model calls")
	}
	if e.Schema != "gooo/workspace-body-execution-receipt/v1" || e.Decision != "OBSERVED" || e.Error != "" || e.Result.Program.Entry.Package != "examples/preview" || e.Result.Program.Entry.Activity != "Plan" {
		return nil, fmt.Errorf("expected an observed preview execution; inspect the saved receipt")
	}
	if r.Calls == nil || *r.Calls != 0 || r.Passed == nil || *r.Passed != 0 || r.Total == nil || *r.Total != 0 || len(r.Traces) != 1 || r.Traces[0].Index != 0 || len(r.Traces[0].Deliveries) != 1 {
		return nil, fmt.Errorf("expected one unscored native execution with zero runtime model calls")
	}
	d := r.Traces[0].Deliveries[0]
	if len(d.Actual) == 0 || string(d.Actual) == "null" || len(d.Expected) != 0 || d.Passed != nil || len(d.Inputs) != 2 || d.Inputs[0].Port != "input0" || d.Inputs[1].Port != "input1" {
		return nil, fmt.Errorf("expected an actual result for the two supplied arguments")
	}
	var title string
	var budget int64
	if json.Unmarshal(d.Inputs[0].Value, &title) != nil || json.Unmarshal(d.Inputs[1].Value, &budget) != nil || title != o.title || budget != o.budget {
		return nil, fmt.Errorf("observed arguments differ from the supplied title or budget")
	}
	return d.Actual, nil
}
