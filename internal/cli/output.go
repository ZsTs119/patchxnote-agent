package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func normalizedOutputFormat(state *rootState) string {
	return strings.ToLower(strings.TrimSpace(state.viper.GetString("output")))
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func unsupportedOutputFormatError(format string) error {
	return fmt.Errorf("unsupported output format %q", format)
}
