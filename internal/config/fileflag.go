package config

import "os"

// FileVar is a custom flag type for files
// This should implement the Value interface of the flag package
// Reference: https://pkg.go.dev/gg-scm.io/tool/internal/flag#FlagSet.Var
type FileVar struct {
	*os.File

	// flag is the flag to open the file with
	// os.O_APPEND|os.O_CREATE|os.O_WRONLY
	flag int
}

// String presents the current value as a string.
func (f *FileVar) String() string {
	if f.File == nil {
		return ""
	}

	return f.Name()
}

// Set is called once, in command line order, for each flag present.
func (f *FileVar) Set(value string) error {
	file, err := os.OpenFile(value, f.flag, 0o644)
	if err != nil {
		return err
	}

	f.File = file
	return nil
}

// Get returns the contents of the Value.
func (f *FileVar) Get() interface{} {
	return f.File
}

// IsBoolFlag returns true if the flag is a boolean flag
func (f *FileVar) IsBoolFlag() bool {
	return false
}
