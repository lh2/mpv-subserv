package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func parseAssTime(timestamp string) float64 {
	var seconds float64
	p := strings.Split(timestamp, ":")
	h, err := strconv.Atoi(p[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "mpv-subserv: ass/h: %v\n", err)
		return -1
	}
	m, err := strconv.Atoi(p[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "mpv-subserv: ass/m: %v\n", err)
		return -1
	}
	s, err := strconv.ParseFloat(p[2], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mpv-subserv: ass/s: %v\n", err)
		return -1
	}
	seconds += s
	seconds += float64(m) * 60
	seconds += float64(h) * 60 * 60
	return seconds
}

func parseAssText(text string) string {
	cleaned := ""
	brc := 0
	for _, c := range []rune(text) {
		switch c {
		case '{':
			brc++
		case '}':
			brc--
		default:
			if brc == 0 {
				cleaned += string(c)
			}
		}
	}
	return strings.TrimSpace(cleaned)
}

func parseAss(file *os.File) ([]LineMsg, error) {
	subs := make([]LineMsg, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") {
			continue
		}
		p := strings.SplitN(line, ":", 2)
		if p[0] != "Dialogue" {
			continue
		}
		p = strings.SplitN(p[1], ",", 10)
		start := parseAssTime(p[1])
		end := parseAssTime(p[2])
		text := parseAssText(p[9])
		if text == "" {
			continue
		}
		msg := LineMsg{
			SubStart: start,
			SubEnd:   end,
			Line:     text,
		}
		fmt.Fprintf(os.Stderr, "%s: %v-%v\n", text, start, end)
		subs = append(subs, msg)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return subs, nil
}
