#!/usr/bin/env bash
# logos-sdk-trial (throwaway): template pre-instantiation prototype.
# BoxInt=lib::Box<int> is declared on the command line; the generator writes
# seaqt_instantiations.h with `extern template class lib::Box<int>;` (parse
# time only), binds the TU-level ClassTemplateSpecializationDecl clang emits
# for it, and lib::Holder's Box<int> members resolve to it.
#
#   OUT=/tmp/out trial/run-templates.sh
set -euo pipefail
here=$(cd "$(dirname "$0")/.." && pwd)
: "${OUT:?}"
(cd "$here" && go build -o "$OUT/genbindings" ./cmd/genbindings)
(cd "$OUT" && SEAQT_GEN_MODULES=Core SEAQT_GEN_LOGOS_SDK="$here/trial/tmpl" \
  SEAQT_GEN_LOGOS_HEADERS=cpp/boxlib.h \
  SEAQT_GEN_INSTANTIATE="BoxInt=lib::Box<int>" SEAQT_GEN_INSTANTIATE_INCLUDES=boxlib.h \
  ./genbindings -clang clang -outdir "$OUT/gen" > "$OUT/genbindings.log" 2>&1)
nimgen=$(echo "$OUT"/gen/nim-seaqt-*)
cp "$here"/libseaqt/seaqt/* "$nimgen/seaqt/"
# -include new: without a Qt header in the chain, the wrappers' std::nothrow is undeclared
sed -i "s|gorge(\"pkg-config --cflags Qt6Core\") \&|gorge(\"pkg-config --cflags Qt6Core\") \& \" -include new -I$here/trial/tmpl/cpp -I$OUT/gen/instantiations\" \&|" \
  "$nimgen/seaqt/LogosSdk/logossdk_pkg.nim"
nim c --hints:off --passL:-lstdc++ --path:"$nimgen" -o:"$OUT/box_demo" "$here/trial/tmpl/box_demo.nim"
"$OUT/box_demo"
