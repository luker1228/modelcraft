package sqlsoftdelete

import (
	"bufio"
	"strings"
)

type QueryBlock struct {
	Header      string
	Body        string
	Annotations Annotations
}

func SplitSQLCFile(src string) []QueryBlock {
	scanner := bufio.NewScanner(strings.NewReader(src))
	blocks := make([]QueryBlock, 0, 8)

	pendingComments := make([]string, 0, 8)
	var current *QueryBlock
	bodyLines := make([]string, 0, 32)

	flushCurrent := func() {
		if current == nil {
			return
		}
		current.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
		blocks = append(blocks, *current)
		current = nil
		bodyLines = bodyLines[:0]
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(strings.ToLower(trimmed), "-- name:") {
			flushCurrent()
			ann := ParseAnnotations([]byte(strings.Join(pendingComments, "\n")))
			current = &QueryBlock{Header: trimmed, Annotations: ann}
			pendingComments = pendingComments[:0]
			continue
		}

		if current != nil {
			bodyLines = append(bodyLines, line)
			continue
		}

		if strings.HasPrefix(trimmed, "--") || trimmed == "" {
			pendingComments = append(pendingComments, line)
		} else {
			pendingComments = pendingComments[:0]
		}
	}

	flushCurrent()

	return blocks
}
