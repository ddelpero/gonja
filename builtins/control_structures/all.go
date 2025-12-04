package controlStructures

import (
	"github.com/ddelpero/gonja/v2/exec"
	"github.com/ddelpero/gonja/v2/parser"
)

var All = exec.NewControlStructureSet(map[string]parser.ControlStructureParser{
	"autoescape": autoescapeParser,
	"block":      blockParser,
	"call":       callParser,
	"extends":    extendsParser,
	"filter":     filterParser,
	"for":        forParser,
	"from":       fromParser,
	"if":         ifParser,
	"import":     importParser,
	"include":    includeParser,
	"macro":      macroParser,
	"raw":        rawParser,
	"set":        setParser,
	"with":       withParser,
})
