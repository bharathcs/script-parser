package scriptparser

import (
	"encoding/csv"
	"io"
)

type Script struct {
	Actors            []Actor
	MapRepresentation map[Actor][]Line
	ArrRepresentation []Line
}

type Actor = string

const NON_ACTOR = Actor("")

type Line struct {
	Speaker     string
	LineNumbers []int
	Dialogue    string
}

type RawLine struct {
	Actor       string
	Line        string
	LineNumbers []int
}

func NewScript(rawDialogues []RawLine) *Script {
	s := Script{MapRepresentation: make(map[Actor][]Line)}
	for _, rawDialogue := range rawDialogues {
		d := Line{
			Speaker:     rawDialogue.Actor,
			LineNumbers: rawDialogue.LineNumbers,
			Dialogue:    rawDialogue.Line,
		}
		_, ok := s.MapRepresentation[rawDialogue.Actor]
		if !ok {
			s.Actors = append(s.Actors, rawDialogue.Actor)
		}
		s.ArrRepresentation = append(s.ArrRepresentation, d)
		actorLines := s.MapRepresentation[rawDialogue.Actor]
		s.MapRepresentation[rawDialogue.Actor] = append(actorLines, d)
	}

	return &s
}

type StringSimplifier = func(string) string
type StringComparator = func(string, string) bool
type SearchFunction = func(string) (bool, Line)

func (s Script) CreateSearchFunction(comparator StringComparator, simplifiers ...StringSimplifier) SearchFunction {
	simplifier := composeSimplifiers(simplifiers)
	var copiedDialogues []Line
	var simplifiedDialogues []string

	copy(copiedDialogues, s.ArrRepresentation)
	for _, dialogue := range copiedDialogues {
		simplifiedDialogues = append(simplifiedDialogues, simplifier(dialogue.Dialogue))
	}

	return func(target string) (bool, Line) {
		for i, simpleDialogue := range simplifiedDialogues {
			if comparator(simpleDialogue, target) {
				return true, copiedDialogues[i]
			}
		}

		return false, Line{}
	}
}

func (s Script) ConvertToCsv(splitAtNewLine bool, wr io.Writer) {
	writer := csv.NewWriter(wr)
	writer.WriteAll(printScript(s, splitAtNewLine))
	writer.Flush()
}

// composeSimplifiers will compose all the functions in the input.
// Returns identity function if the input is nil.
// Note: it is right associative e.g. ( f, g, h ) => ( x => f(g(h(x)) )
func composeSimplifiers(simplifiers []StringSimplifier) StringSimplifier {
	switch len(simplifiers) {
	case 0:
		return func(s string) string { return s }
	case 1:
		return func(s string) string { return simplifiers[0](s) }
	default:
		return func(s string) string {
			return simplifiers[0](composeSimplifiers(simplifiers[1:])(s))
		}
	}
}
