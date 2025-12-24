package domain

import (
	"encoding/json"
)

/*
	_index:   key=streamID, value=version
	_content: key=streamID, value=map[streamPos]globalPos
	_entries: key=globalPos, value=event
*/

type Index struct {
	streamID string
	version  uint64
}

type Content struct {
	streamID  string
	positions map[uint64]uint64
}

type Entries struct {
	globalPos uint64
	events    Event
}

func (c *Content) MarshalJSON() ([]byte, error) {
	v := map[string]any{
		"StreamID":  c.streamID,
		"Positions": c.positions,
	}
	for streamPos, globalPos := range c.positions {
		v["Positions"].(map[uint64]uint64)[streamPos] = globalPos
	}
	return json.MarshalIndent(v, "", "  ")
}

func (c *Content) UnmarshalJSON(data []byte) error {
	var v map[string]any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	c.streamID = v["StreamID"].(string)
	var pmap map[uint64]uint64
	pmap = v["Positions"].(map[uint64]uint64)
	c.positions = make(map[uint64]uint64, len(pmap))
	for k, v := range pmap {
		c.positions[k] = v
	}
	return nil
}
