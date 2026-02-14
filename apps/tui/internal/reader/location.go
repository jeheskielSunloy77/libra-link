package reader

import (
	"fmt"
	"strconv"
	"strings"
)

func EncodeLocation(doc *Document, line int) string {
	if doc == nil {
		return "fmt=txt;line=0"
	}
	line = ClampLocation(doc, line)
	format := strings.ToLower(strings.TrimSpace(doc.Format))
	if format == "" {
		format = "txt"
	}

	anchor := lineAnchorAt(doc, line)
	switch format {
	case "pdf":
		page := anchor.Page
		if page <= 0 {
			page = inferPageForLine(doc, line)
		}
		if page <= 0 {
			page = 1
		}
		return fmt.Sprintf("fmt=pdf;page=%d;line=%d", page, line)
	case "epub":
		spine := anchor.Spine
		if spine < 0 {
			spine = inferSpineForLine(doc, line)
		}
		offset := anchor.Offset
		if offset < 0 {
			offset = line
		}
		return fmt.Sprintf("fmt=epub;spine=%d;offset=%d;line=%d", spine, offset, line)
	default:
		return fmt.Sprintf("fmt=txt;line=%d", line)
	}
}

func DecodeLocation(doc *Document, token string) (int, bool) {
	if doc == nil {
		return 0, false
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, false
	}

	if strings.HasPrefix(token, "line:") {
		line, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(token, "line:")))
		if err != nil {
			return 0, false
		}
		return ClampLocation(doc, line), true
	}

	fields := parseLocationFields(token)
	line, hasLine, lineErr := parseIntField(fields, "line")
	if hasLine && lineErr == nil {
		return ClampLocation(doc, line), true
	}

	fmtType := strings.ToLower(strings.TrimSpace(fields["fmt"]))
	switch fmtType {
	case "pdf":
		page, ok, err := parseIntField(fields, "page")
		if ok && err == nil {
			if line, found := findLineByPage(doc, page); found {
				return line, true
			}
		}
	case "epub":
		spine, hasSpine, err := parseIntField(fields, "spine")
		if hasSpine && err == nil {
			offset, hasOffset, offsetErr := parseIntField(fields, "offset")
			if hasOffset && offsetErr == nil {
				if line, found := findLineBySpineOffset(doc, spine, offset); found {
					return line, true
				}
			}
			if line, found := findFirstLineBySpine(doc, spine); found {
				return line, true
			}
		}
	}

	return 0, false
}

func parseLocationFields(token string) map[string]string {
	out := map[string]string{}
	for _, pair := range strings.Split(token, ";") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		if key != "" {
			out[key] = value
		}
	}
	return out
}

func parseIntField(fields map[string]string, key string) (int, bool, error) {
	raw, ok := fields[key]
	if !ok {
		return 0, false, nil
	}
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, true, err
	}
	return value, true, nil
}

func lineAnchorAt(doc *Document, line int) LineAnchor {
	if doc == nil || line < 0 || line >= len(doc.LineIndex) {
		return LineAnchor{Line: line, Page: -1, Spine: -1, Offset: -1}
	}
	return doc.LineIndex[line]
}

func inferPageForLine(doc *Document, line int) int {
	if doc == nil {
		return -1
	}
	for i := line; i >= 0 && i < len(doc.LineIndex); i-- {
		if doc.LineIndex[i].Page > 0 {
			return doc.LineIndex[i].Page
		}
	}
	for i := line + 1; i < len(doc.LineIndex); i++ {
		if doc.LineIndex[i].Page > 0 {
			return doc.LineIndex[i].Page
		}
	}
	return -1
}

func inferSpineForLine(doc *Document, line int) int {
	if doc == nil {
		return -1
	}
	for i := line; i >= 0 && i < len(doc.LineIndex); i-- {
		if doc.LineIndex[i].Spine >= 0 {
			return doc.LineIndex[i].Spine
		}
	}
	for i := line + 1; i < len(doc.LineIndex); i++ {
		if doc.LineIndex[i].Spine >= 0 {
			return doc.LineIndex[i].Spine
		}
	}
	return -1
}

func findLineByPage(doc *Document, page int) (int, bool) {
	if doc == nil {
		return 0, false
	}
	for i, anchor := range doc.LineIndex {
		if anchor.Page == page {
			return ClampLocation(doc, i), true
		}
	}
	return 0, false
}

func findLineBySpineOffset(doc *Document, spine, offset int) (int, bool) {
	if doc == nil {
		return 0, false
	}
	for i, anchor := range doc.LineIndex {
		if anchor.Spine == spine && anchor.Offset == offset {
			return ClampLocation(doc, i), true
		}
	}
	return 0, false
}

func findFirstLineBySpine(doc *Document, spine int) (int, bool) {
	if doc == nil {
		return 0, false
	}
	for i, anchor := range doc.LineIndex {
		if anchor.Spine == spine {
			return ClampLocation(doc, i), true
		}
	}
	return 0, false
}
