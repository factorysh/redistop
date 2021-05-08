package monitor

import "strings"

func BulkTable(bulk string) (map[string]string, error) {
	s := make(map[string]string)
	lines := strings.Split(bulk, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		values := strings.Split(line, ":")
		if len(values) > 1 {
			s[values[0]] = values[1][:len(values[1])]
		}
	}
	return s, nil
}
