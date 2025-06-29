package utils

import (
	"fmt"
	"io"
	"encoding/json"
)

// Helper functions for SSE formatting
func WriteSSEData(w io.Writer, data string) {
	fmt.Fprintf(w, "data: %s\n\n", data)
}

func WriteSSEMessage(w io.Writer, message string) {
	data := map[string]string{"message": message}
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(w, "data: {\"message\": \"Error creating message\"}\n\n")
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
}