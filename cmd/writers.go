package cmd

import (
	"io"
	"os"
)

// orStdout returns w when set, otherwise os.Stdout. Services built directly in
// tests leave the writer nil.
func orStdout(w io.Writer) io.Writer {
	if w != nil {
		return w
	}
	return os.Stdout
}

// orStderr returns w when set, otherwise os.Stderr. Services built directly in
// tests leave the writer nil.
func orStderr(w io.Writer) io.Writer {
	if w != nil {
		return w
	}
	return os.Stderr
}
