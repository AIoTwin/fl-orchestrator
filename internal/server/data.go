package server

import (
	"encoding/json"
	"io"

	"github.com/AIoTwin-Adaptive-FL-Orch/fl-orchestrator/internal/florch/cost"
)

func toJSON(i interface{}, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(i)
}

func fromJSON(i interface{}, r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(i)
}

type StartFlRequest struct {
	Epochs             int32                 `json:"epochs"`
	LocalRounds        int32                 `json:"localRounds"`
	ModelSize          float32               `json:"modelSize"`
	CostConfiguration  cost.CostCofiguration `json:"costConfiguration" `
	ConfigurationModel string                `json:"configurationModel"`
	RvaEnabled         bool                  `json:"rvaEnabled"`
}
