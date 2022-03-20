package scriptparser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
)

type ScriptParser struct {
	actorDialoguePattern *regexp.Regexp
	isMultiLineDialogue  bool
	nonDialogueRegexp    []*regexp.Regexp
}

// CreateDialogueRegex creates a regexp pattern for identifying the start of a dialogue.
// Note that it will be combined in the form: `^{prefixRegexp}{actorRegexp}{postfixRegexp}{dialogue}`.
// If isMultiString is turned on, dialogue will include every line afterwards until
// `^{prefixRegexp}{actorRegexp}{postfixRegexp}` is found on the next line. Do not use capturing groups inside the input
// regexp. Note: any errors will cause a panic to simplify client code.
func (s *ScriptParser) CreateDialogueRegex(prefixRegexp, actorRegexp, postfixRegexp string, isMultiLineDialogue bool) {
	capturingGroup := regexp.MustCompile(`\(\?P<.*>\)`)
	for _, v := range []string{prefixRegexp, actorRegexp, postfixRegexp} {
		if capturingGroup.MatchString(v) {
			panic(fmt.Errorf("regexp inputs should not have capture groups: %q", v))
		}
		if _, subExprErr := regexp.Compile(v); subExprErr != nil {
			panic(fmt.Errorf("regexp input did not compile: %q (%w)", v, subExprErr))
		}
		if _, subExprErr := regexp.Compile(v); subExprErr != nil {
			panic(fmt.Errorf("regexp input did not compile: %q (%w)", v, subExprErr))
		}
	}

	fullRegexp := `^` + prefixRegexp + `(?P<Actor>` + actorRegexp + `)` + postfixRegexp + `(?P<Line>.*)`
	if fullRegexp, err := regexp.Compile(fullRegexp); err != nil {
		panic(fmt.Errorf("full regexp did not compile: %q (%w)", fullRegexp, err))
	} else {
		s.actorDialoguePattern = fullRegexp
		s.isMultiLineDialogue = isMultiLineDialogue
	}
}

// UseSkipRegexps to skip over non-dialogue lines if isMultilineDialog is turned on.
// Unlike CreateDialogRegex, this is not combined with anything else and will be used as is. As such, the regex should
// use ^ and $ to ensure multiline dialogs that contain a match do not get skipped over.
// Note: Only has an effect if isMultilineDialog is true.
func (s *ScriptParser) UseSkipRegexps(skipRegexps []*regexp.Regexp) {
	s.nonDialogueRegexp = skipRegexps
}

// LoadTranscriptFromFilePath will attempt to load from filepath, and panic immediately if it does not work.
func (s *ScriptParser) LoadTranscriptFromFilePath(filepath string) {
	f, err := os.Open(filepath)
	if err != nil {
		panic(fmt.Sprintf("Unable to open filepath %q: %s\n", filepath, err.Error()))
	}
	s.LoadTranscript(f)
}

func (s *ScriptParser) LoadTranscript(reader io.Reader) *Script {
	scanner := bufio.NewScanner(reader)

	var rawScript []RawLine
	var prevLine RawLine
	currentLineNumber := -1

	for scanner.Scan() {
		currentLineNumber++
		line := scanner.Text()

		if isNewLine, newLine := createNewRawLineIfMatch(s.actorDialoguePattern, line, currentLineNumber); isNewLine {
			rawScript = append(rawScript, prevLine)
			prevLine = newLine

		} else if prevLine.LineNumbers == nil || !s.isMultiLineDialogue || isNonDialogue(s.nonDialogueRegexp, line) {
			/*
				Create a new RawLine if any of the following is true:
				- is the first line (prevLine.LineNumbers will be an empty array)
				- multiline flag not set (definitely not part of previous line)
				- if it matches with any of the nonDialogue regexps set by user
			*/

			rawScript = append(rawScript, prevLine)
			prevLine = RawLine{
				Actor:       NON_ACTOR,
				Line:        line,
				LineNumbers: []int{currentLineNumber},
			}
		} else {
			// Combine with prevLine because:
			// not the first line of transcript AND isMultilineDialog AND not a nonDialogue line
			prevLine.Line += "\n" + line
			prevLine.LineNumbers = append(prevLine.LineNumbers, currentLineNumber)
		}
	}

	rawScript = append(rawScript, prevLine) // add in last prevLine
	rawScript = rawScript[1:]               // ignore the first prevLine as it is always blank

	if err := scanner.Err(); err != nil {
		panic(fmt.Errorf("unable to scan the entire transcript (stopped at %d): %w", currentLineNumber+1, err))
	}

	return NewScript(rawScript)
}

func createNewRawLineIfMatch(actorDialoguePattern *regexp.Regexp, line string, currentLineNumber int) (bool, RawLine) {
	if !actorDialoguePattern.MatchString(line) {
		return false, RawLine{}
	}

	matches := actorDialoguePattern.FindStringSubmatch(line)
	actor := matches[1]
	lines := matches[2]
	return true, RawLine{
		Actor:       actor,
		Line:        lines,
		LineNumbers: []int{currentLineNumber},
	}
}

// isNonDialogue returns true if line matches any of the skipRegexps
func isNonDialogue(skipRegexps []*regexp.Regexp, line string) bool {
	for _, skipRegexp := range skipRegexps {
		if skipRegexp.MatchString(line) {
			return true
		}
	}
	return false
}
