package main

// Template pre-instantiation, prototype only (logos-sdk-trial).
//
// A configured list of instantiations ("lib::Box<int>" named "BoxInt") is
// written into a generated header as `extern template class lib::Box<int>;`
// lines, guarded by SEAQT_GEN_PARSE so they exist only while clang parses.
// An explicit instantiation *declaration* makes clang emit a TU-level
// ClassTemplateSpecializationDecl, located on that line of the generated
// header, whose member declarations carry the substituted types - without
// instantiating any member definition (so members that would not compile
// for T are not an error at parse time).
//
// The class keeps its C++ spelling as ClassName ("lib::Box<int>"), so every
// C++ call site is valid as-is; cabiClassName maps it to the configured C
// name, and the existing inner-class typedef workaround covers the forward
// declaration in the C header.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type trialInstantiation struct {
	Cpp   string // lib::Box<int>
	Alias string // BoxInt
	Line  int    // line of the extern template in the generated header
}

var (
	trialInstFile  string
	trialInstByLn  = map[int]trialInstantiation{}
	trialCabiAlias = map[string]string{}
)

// writeTrialInstantiations writes the generated header and registers the
// spellings clang uses for each instantiation.
func writeTrialInstantiations(dir string, includes []string, spec string) string {
	var b strings.Builder
	b.WriteString("#pragma once\n")
	for _, inc := range includes {
		fmt.Fprintf(&b, "#include <%s>\n", inc)
	}
	b.WriteString("#ifdef SEAQT_GEN_PARSE\n")
	line := strings.Count(b.String(), "\n") + 1
	for _, item := range strings.Split(spec, ";") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, "=", 2)
		alias, cpp := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		fmt.Fprintf(&b, "extern template class %s;\n", cpp)
		trialInstByLn[line] = trialInstantiation{cpp, alias, line}
		trialCabiAlias[cpp] = alias
		// clang spells the type unqualified inside its own namespace
		if i := strings.LastIndex(cpp[:strings.Index(cpp, "<")], "::"); i >= 0 {
			short := cpp[i+2:]
			KnownTypedefs[short] = lookupResultTypedef{"trial", "trial", CppTypedef{short, parseSingleTypeString(cpp)}}
		}
		line++
	}
	b.WriteString("#endif\n")
	os.MkdirAll(dir, 0755)
	trialInstFile = filepath.Join(dir, "seaqt_instantiations.h")
	if err := os.WriteFile(trialInstFile, []byte(b.String()), 0644); err != nil {
		panic(err)
	}
	return trialInstFile
}
