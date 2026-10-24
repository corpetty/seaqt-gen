# Throwaway: Nim reaching the Logos C++ SDK's non-virtual accessors through
# seaqt-gen output (branch logos-sdk-trial).
import seaqt/QtCore/gen_qcoreapplication
import seaqt/LogosSdk/[gen_token_manager, gen_logos_api]

let app = QCoreApplication.create()
let tm = TokenManager.instance()
tm.saveToken("capability_module", "tok-123")
doAssert tm.getToken("capability_module") == "tok-123"

# What a plugin gets from the host: a LogosAPI*. Its token manager is a
# non-virtual, out-of-line accessor - exactly what blocked option A.
let api = LogosAPI.create("muster_ui")
let viaApi = api.getTokenManager()
doAssert viaApi.h == tm.h, "same singleton"
echo "token via LogosAPI*: ", viaApi.getToken("capability_module")
echo "keys: ", viaApi.getTokenKeys()
