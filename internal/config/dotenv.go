package config

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
)

func loadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || os.Getenv(key) != "" {
			continue
		}

		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		}

		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}

	return scanner.Err()
}
