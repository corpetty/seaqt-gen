import seaqt/LogosSdk/[gen_seaqt_instantiations, gen_boxlib]
let b = BoxInt.make(41)
var v = cint(42)
b.set(addr v)
echo "BoxInt.get: ", b.get()
let h = libHolder.create()
echo "Holder.box().get(): ", h.box().get()
