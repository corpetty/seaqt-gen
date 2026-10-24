package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// selectModules narrows the module list to $SEAQT_GEN_MODULES (comma-separated, e.g.
// "Core,Network,RemoteObjects"), keeping the list's order. A module's dependencies must be
// named too: types are only known once their module has been parsed, and a method using an
// unknown type is dropped silently. Unset = every module.
func selectModules(modules []string) []string {
	want := os.Getenv("SEAQT_GEN_MODULES")
	if want == "" {
		return modules
	}
	keep := map[string]bool{}
	for _, m := range strings.Split(want, ",") {
		keep[strings.TrimSpace(m)] = true
	}
	var ret []string
	for _, m := range modules {
		if keep[m] {
			ret = append(ret, m)
			delete(keep, m)
		}
	}
	for m := range keep {
		log.Panicf("SEAQT_GEN_MODULES names %q, which is not in the module list", m)
	}
	log.Printf("SEAQT_GEN_MODULES: generating %v", ret)
	return ret
}

func toMajor(qtVersion string) string {
	parts := strings.SplitN(qtVersion, ".", 2)
	return parts[0]
}
func toMajorMinor(qtVersion string) string {
	parts := strings.SplitN(qtVersion, ".", 3)
	return parts[0] + "." + parts[1]
}

func ProcessLibraries(clangBin, outDir, extraLibsDir string) {
	qtVersion := qmakeQuery("QT_VERSION")
	major := toMajor(qtVersion)
	majorMinor := toMajorMinor(qtVersion)
	CacheSubdir = "qt" + majorMinor

	if major == "5" {
		buildQt5(clangBin, outDir, extraLibsDir, qtVersion)
	} else if major == "6" {
		buildQt6(clangBin, outDir, extraLibsDir, qtVersion)
	} else {
		panic("unsupported qt version: " + major)
	}
}

func buildQt5(clangBin, outDir, extraLibsDir, qtVersion string) {
	flushKnownTypes()
	InsertTypedefs(false)

	qtHeaders := qmakeQuery("QT_INSTALL_HEADERS")
	majorMinor := toMajorMinor(qtVersion)
	qtDir := filepath.Join(outDir, "seaqt-"+majorMinor)
	seaqtDir := filepath.Join(qtDir, "seaqt")
	nimSeaqtDir := filepath.Join(outDir, "nim-seaqt-"+majorMinor, "seaqt")
	nimbleDir := filepath.Join(outDir, "nim-seaqt-"+majorMinor)

	os.RemoveAll(seaqtDir)
	os.RemoveAll(nimSeaqtDir)

	writeNimbleFile(nimbleDir, qtVersion)

	modules := []string{
		// https://doc.qt.io/qt-5/qtmodules.html#qt-essentials
		"Core", "Gui", "Widgets", "Network", "Multimedia", "MultimediaWidgets",
		"Sql",

		// https://doc.qt.io/qt-5/qtmodules.html#qt-add-ons
		"PrintSupport", "Script", "Svg", "WebChannel", "WebSockets",
		"WebEngineCore", "WebEngineWidgets",

		// OpenGL is deprecated

		"Pdf", "PdfWidgets", "Positioning",

		"Qml", "Quick",
		// TODO Not sure where these are from :)
		"WebKit", "WebKitWidgets",

		// Tools
		// TODO(build fails) "Designer", "UiPlugin", "UiTools",
	}
	for _, lib := range modules {
		generate(
			"Qt"+lib, "Qt5"+lib, qtHeaders, qtVersion,
			[]string{filepath.Join(qtHeaders, "Qt"+lib)},
			clangBin, "", seaqtDir, nimSeaqtDir,
		)
	}
}

func buildQt6(clangBin, outDir, extraLibsDir, qtVersion string) {
	flushKnownTypes()
	InsertTypedefs(true)

	qtHeaders := qmakeQuery("QT_INSTALL_HEADERS")
	majorMinor := toMajorMinor(qtVersion)
	qtDir := filepath.Join(outDir, "seaqt-"+majorMinor)
	seaqtDir := filepath.Join(qtDir, "seaqt")
	nimSeaqtDir := filepath.Join(outDir, "nim-seaqt-"+majorMinor, "seaqt")
	nimbleDir := filepath.Join(outDir, "nim-seaqt-"+majorMinor)

	os.RemoveAll(seaqtDir)
	os.RemoveAll(nimSeaqtDir)

	writeNimbleFile(nimbleDir, qtVersion)

	modules := []string{
		"Core", "Gui", "Widgets", "Network", "Multimedia", "MultimediaWidgets",
		"Sql",

		"PrintSupport", "Svg", "SvgWidgets", "SpatialAudio", "WebChannel", "WebSockets",
		"WebEngineCore", "WebEngineWidgets",

		"Pdf", "PdfWidgets", "Positioning",

		"StateMachine",

		"Designer", "UiPlugin", "UiTools",

		"Qml",

		"Quick", "WebEngineQuick",
	}

	for _, lib := range selectModules(modules) {
		generate(
			"Qt"+lib, "Qt6"+lib, qtHeaders, qtVersion,
			[]string{filepath.Join(qtHeaders, "Qt"+lib)},
			clangBin, "--std=c++17", seaqtDir, nimSeaqtDir,
		)
	}

	// logos-sdk-trial (throwaway): the Logos C++ SDK headers a plugin needs, as a
	// "library" that depends on QtCore. SEAQT_GEN_LOGOS_SDK=<logos-cpp-sdk checkout>,
	// SEAQT_GEN_LOGOS_HEADERS=comma-separated list relative to it.
	if sdk := os.Getenv("SEAQT_GEN_LOGOS_SDK"); sdk != "" {
		var hdrs []string
		for _, h := range strings.Split(os.Getenv("SEAQT_GEN_LOGOS_HEADERS"), ",") {
			hdrs = append(hdrs, filepath.Join(sdk, strings.TrimSpace(h)))
		}
		extra := ""
		// SEAQT_GEN_INSTANTIATE="BoxInt=lib::Box<int>;..." with
		// SEAQT_GEN_INSTANTIATE_INCLUDES="box.h,...": template pre-instantiation trial
		if spec := os.Getenv("SEAQT_GEN_INSTANTIATE"); spec != "" {
			incs := strings.Split(os.Getenv("SEAQT_GEN_INSTANTIATE_INCLUDES"), ",")
			hdrs = append(hdrs, writeTrialInstantiations(filepath.Join(outDir, "instantiations"), incs, spec))
			extra = " -DSEAQT_GEN_PARSE -I" + filepath.Join(outDir, "instantiations")
		}
		generate(
			"LogosSdk", "Qt6Core", qtHeaders, qtVersion, hdrs,
			clangBin, "--std=c++17 -I"+filepath.Join(sdk, "cpp")+" -I"+filepath.Join(sdk, "core")+extra,
			seaqtDir, nimSeaqtDir,
		)
	}
}
