package output

import "io"

func WriteSuccess(w io.Writer, format string, compact bool, data any, meta map[string]any) error {
	if err := validateFormat(format); err != nil {
		return err
	}

	payload := struct {
		OK   bool           `json:"ok"`
		Data any            `json:"data"`
		Meta map[string]any `json:"meta,omitempty"`
	}{
		OK:   true,
		Data: data,
		Meta: meta,
	}
	return writeJSON(w, compact, payload)
}
