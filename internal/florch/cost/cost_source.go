package cost

import (
	"encoding/json"
	"fmt"
	"strings"
)

type CostSource int

const (
	ENERGY CostSource = iota
	COMMUNICATION
)

func (c CostSource) String() string {
	switch c {
	case ENERGY:
		return "ENERGY"
	case COMMUNICATION:
		return "COMMUNICATION"
	default:
		return "UNKNOWN"
	}
}

// Marshal as a JSON string: "ENERGY"/"COMMUNICATION"
func (c CostSource) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

// Accept either JSON strings ("ENERGY") or numbers (0/1)
func (c *CostSource) UnmarshalJSON(b []byte) error {
	// string path
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		s := strings.Trim(string(b), `"`)
		switch strings.ToUpper(s) {
		case "ENERGY":
			*c = ENERGY
		case "COMMUNICATION":
			*c = COMMUNICATION
		default:
			return fmt.Errorf("invalid CostSource: %q", s)
		}
		return nil
	}
	// numeric path
	var i int
	if err := json.Unmarshal(b, &i); err != nil {
		return err
	}
	switch v := CostSource(i); v {
	case ENERGY, COMMUNICATION:
		*c = v
		return nil
	default:
		return fmt.Errorf("invalid CostSource numeric value: %d", i)
	}
}
