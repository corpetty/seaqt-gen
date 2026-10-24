#!/usr/bin/env bash
# logos-sdk-trial (throwaway, not for merge): seaqt-gen on the Logos C++ SDK headers.
#
# Needs: clang 14.0.6 (with -fPIC; Qt 6 built with reduce_relocations refuses
# otherwise), go, Qt 6 (qmake + pkg-config, incl. Qt6RemoteObjects for the SDK),
# nim 2.2, and a liblogos_sdk.a built from the same SDK checkout and Qt.
#
#   SDK=~/logos-cpp-sdk OUT=/tmp/out LOGOS_SDK_LIB=/path/liblogos_sdk.a trial/run.sh
set -euo pipefail
here=$(cd "$(dirname "$0")/.." && pwd)
: "${SDK:?}" "${OUT:?}" "${LOGOS_SDK_LIB:?}"
(cd "$here" && go build -o "$OUT/genbindings" ./cmd/genbindings)
(cd "$OUT" && SEAQT_GEN_MODULES=Core SEAQT_GEN_LOGOS_SDK="$SDK" \
  SEAQT_GEN_LOGOS_HEADERS=cpp/logos_api.h,cpp/token_manager.h,cpp/logos_api_client.h,cpp/logos_object.h,cpp/logos_types.h,cpp/logos_mode.h,cpp/logos_api_provider.h,cpp/logos_provider_object.h,core/interface.h \
  ./genbindings -clang clang -outdir "$OUT/gen" > "$OUT/genbindings.log" 2>&1)
nimgen=$(echo "$OUT"/gen/nim-seaqt-*)
cp "$here"/libseaqt/seaqt/* "$nimgen/seaqt/"
# What a library configuration would have to express, and the generator cannot
# today: the library's own include path and link line in the emitted *_pkg.nim
# (it only knows one pkg-config name per module).
sed -i "s|gorge(\"pkg-config --cflags Qt6Core\") \&|gorge(\"pkg-config --cflags Qt6Core\") \& \" -I$SDK/cpp -I$SDK/core\" \&|; s|LogosSdkLibs\* = gorge(\"pkg-config --libs Qt6Core\")|LogosSdkLibs* = \"$LOGOS_SDK_LIB \" \& gorge(\"pkg-config --libs Qt6Core Qt6RemoteObjects\")|" \
  "$nimgen/seaqt/LogosSdk/logossdk_pkg.nim"
nim c --hints:off --passL:-lstdc++ --path:"$nimgen" -o:"$OUT/token_demo" "$here/trial/token_demo.nim"
QT_QPA_PLATFORM=offscreen "$OUT/token_demo"
