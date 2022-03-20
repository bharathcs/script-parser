package scriptparser

import (
	"bufio"
	"fmt"
	"strings"
)

type lineRepresentation struct {
	speaker         string
	line            string
	firstLineNumber int
	dialogueId      int
}

func printScript(s Script, splitAtNewLine bool) [][]string {
	output := [][]string{
		{"Dialogue ID", "Line Number", "Speaker", "Dialogue"},
	}

	for dialogueId, line := range s.ArrRepresentation {
		if splitAtNewLine {
			scanner := bufio.NewScanner(strings.NewReader(line.Dialogue))
			for scanner.Scan() {
				lineConverted := createLineRepresentation(scanner.Text(), line.Speaker, dialogueId, line.LineNumbers[0])
				output = append(output, lineConverted.toStrings())
			}
		} else {
			lineConverted := createLineRepresentation(line.Dialogue, line.Speaker, dialogueId, line.LineNumbers[0])
			output = append(output, lineConverted.toStrings())
		}
		output = append(output)
	}
	return output
}

func createLineRepresentation(line, speaker string, dialogueId int, lineNumber int) lineRepresentation {
	return lineRepresentation{
		speaker:         speaker,
		line:            line,
		firstLineNumber: lineNumber,
		dialogueId:      dialogueId,
	}
}

func (l lineRepresentation) toStrings() []string {
	return []string{fmt.Sprint(l.dialogueId), fmt.Sprint(l.firstLineNumber), l.speaker, l.line}
}
