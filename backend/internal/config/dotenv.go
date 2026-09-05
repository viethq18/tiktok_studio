package config

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// LoadDotEnv walks up from the working directory looking for a .env file and
// exports any key that is not already set in the environment.
func LoadDotEnv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for i := 0; i < 5; i++ {
		path := filepath.Join(dir, ".env")
		if f, err := os.Open(path); err == nil {
			defer f.Close()
			sc := bufio.NewScanner(f)
			for sc.Scan() {
				line := strings.TrimSpace(sc.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				k, v, ok := strings.Cut(line, "=")
				if !ok {
					continue
				}
				k, v = strings.TrimSpace(k), strings.Trim(strings.TrimSpace(v), `"'`)
				if _, exists := os.LookupEnv(k); !exists {
					os.Setenv(k, v)
				}
			}
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}
