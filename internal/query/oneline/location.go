package oneline

import "encoding/json"

func Location(documentPath string, scipRange []int32) (string, int32, int32) {
	if documentPath == "" {
		return "?", 0, 0
	}
	if len(scipRange) < 2 {
		return documentPath, 0, 0
	}

	return documentPath, scipRange[0] + 1, scipRange[1] + 1
}

func Quote(value string) string {
	quoted, err := json.Marshal(value)
	if err != nil {
		panic("quote one-line field: " + err.Error())
	}

	return string(quoted)
}
