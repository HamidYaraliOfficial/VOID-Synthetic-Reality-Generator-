package main

import "strconv"

// flags is a minimal `--key value` parser (no external CLI framework
// needed) supporting string/int/int64 flags with defaults.
type flags struct {
	m map[string]string
}

func parseFlags(args []string) *flags {
	f := &flags{m: map[string]string{}}
	for i := 0; i < len(args); i++ {
		a := args[i]
		if len(a) > 2 && a[0] == '-' && a[1] == '-' {
			key := a[2:]
			if i+1 < len(args) {
				f.m[key] = args[i+1]
				i++
			} else {
				f.m[key] = "true"
			}
		}
	}
	return f
}

func (f *flags) str(key, def string) string {
	if v, ok := f.m[key]; ok {
		return v
	}
	return def
}

func (f *flags) int(key string, def int) int {
	if v, ok := f.m[key]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func (f *flags) int64(key string, def int64) int64 {
	if v, ok := f.m[key]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}
