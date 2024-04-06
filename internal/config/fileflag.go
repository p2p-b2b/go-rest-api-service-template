package config

import "os"

// FileFlag is a custom flag type for files
// Reference: https://pkg.go.dev/gg-scm.io/tool/internal/flag#FlagSet.Var
type FileFlag struct {
	*os.File
}

// String presents the current value as a string.
func (f *FileFlag) String() string {
	if f.File == nil {
		return ""
	}

	return f.File.Name()
}

// Set is called once, in command line order, for each flag present.
func (f *FileFlag) Set(value string) error {
	file, err := os.OpenFile(value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}

	f.File = file
	return nil
}

// Get returns the contents of the Value.
func (f *FileFlag) Get() interface{} {
	return f.File
}

// IsBoolFlag returns true if the flag is a boolean flag
func (f *FileFlag) IsBoolFlag() bool {
	return false
}
