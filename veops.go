package main

/*
	assembler instructions as of VE 1.0
*/

var bcodes = map[string]string{
	"af":    "Always false",
	"gt":    "Greater than",
	"lt":    "Less than",
	"ne":    "Not equal",
	"eq":    "Equal",
	"ge":    "Greater than or equal",
	"le":    "Less than or equal",
	"num":   "Is number",
	"nan":   "Is NaN (Not a number)",
	"gtnan": "Greater than or NaN",
	"ltnan": "Less then or NaN",
	"nenan": "Not equal or NaN",
	"eqnan": "Equal or NaN",
	"genan": "Greater than equal or NaN",
	"lenan": "Greater than equal or NaN",
	"at":    "Always true",
	"":      "Always true"}

var suffixes = map[string]string{
	".l":     "64bit integer",
	".w":     "32bit integer",
	".d":     "64bit floating point",
	".s":     "32bit floating point",
	".sx":    "sign extension",
	".zx":    "zero extension",
	".nt":    "branch not taken",
	".t":     "branch taken",
	".af":    "Always false",
	".gt":    "Greater than",
	".lt":    "Less than",
	".ne":    "Not equal",
	".eq":    "Equal",
	".ge":    "Greater than or equal",
	".le":    "Less than or equal",
	".num":   "Is number",
	".nan":   "Is NaN (Not a number)",
	".gtnan": "Greater than or NaN",
	".ltnan": "Less then or NaN",
	".nenan": "Not equal or NaN",
	".eqnan": "Equal or NaN",
	".genan": "Greater than equal or NaN",
	".lenan": "Greater than equal or NaN",
	".at":    "Always true",
	".rz":    "round towards zero",
	".rp":    "round towards plus infinity",
	".rm":    "round towards minus infinity",
	".rn":    "round to nearest (ties to even)",
	".ra":    "round to nearest (ties to away)",
	".nc":    "not cached", // FIXME correct?
	".ot":    "overtake",
	".nex":   "no exception",
	".fst":   "first element",
	".lst":   "last element"}

var registers = map[string]string{
	"%sp":    "stack pointer (%s11)",
	"%fp":    "frame pointer (%s9)",
	"%sl":    "stack limit (%s8)",
	"%lr":    "link register (%s10)",
	"%tp":    "thread pointer (%s14)",
	"%outer": "outer register (%s12)",
	"%info":  "info area register (%s17)",
	"%got":   "global offset table (%s15)",
	"%plt":   "procedure linkage table (%s16)",
	"%usrcc": "user clock counter",
	"%psw":   "program status word",
	"%sar":   "store address register",
	"%pmmr":  "performance monitor mode register",
	"%pmcr":  "performance monitor configuration register",
	"%pmc":   "performance monitor counter"}

var veops = [...]string{
	".local symbol = Local Symbol",
	".balign bytes = Align here",
	".type ",
	".size ",
	".zero ",
	".data = Data segment",
	".text = Text Segment",
	".bss = BSS Segment",
	".section name[, \"flags\"[, @type[, flag_specific_arguments]]]",
	".text [subsection] = Text segment",
	".data [subsection] = Data segment",
	".bss [.subsection] = BSS Segment",
	".cfi_startproc",
	".cfi_endproc",
	".2byte EXPRESSIONS = These directives write unaligned 2, 4 or 8 byte values to the output section.",
	".4byte EXPRESSIONS = These directives write unaligned 2, 4 or 8 byte values to the output section.",
	".8byte EXPERSSIONS = These directives write unaligned 2, 4 or 8 byte values to the output section.",
	".byte EXPRESSIONS = .byte expects zero or more expressions, separated by commas. ",
	".short EXPRESSIONS = Expect zero or more EXPRESSIONS, of any section, separated by commas.  For each expression, emit a number that, at run time, is the value of that expression. The byte order is little endian and bit size of the number is 16 bits (2 bytes).",
	".word EXPRESSIONS = Expect zero or more EXPRESSIONS, of any section, separated by commas.  For each expression, emit a number that, at run time, is the value of that expression. The byte order is little endian and bit size of the number is 32 bits (4 bytes).",
	".int EXPRESSIONS = Expect zero or more EXPRESSIONS, of any section, separated by commas.  For each expression, emit a number that, at run time, is the value of that expression. The byte order is little endian and bit size of the number is 32 bits (4 bytes).",
	".long EXPRESSIONS = Expect zero or more EXPRESSIONS, of any section, separated by commas.  For each expression, emit a number that, at run time, is the value of that expression. The byte order is little endian and bit size of the number is 64 bits (8 bytes).",
	".quad EXPRESSIONS = Expect zero or more EXPRESSIONS, of any section, separated by commas.  For each expression, emit a number that, at run time, is the value of that expression. The byte order is little endian and bit size of the number is 64 bits (8 bytes).",
	".llong EXPRESSIONS = Expect zero or more EXPRESSIONS, of any section, separated by commas.  For each expression, emit a number that, at run time, is the value of that expression. The byte order is little endian and bit size of the number is 64 bits (8 bytes).",
	".file \"string\" = source file name",
	".ident \"string\" = string is emitted to the \".comment\" section.",
	"lea %sx, ASX = Load Effective Address",
	"lea.sl %sx, ASX = Load Effective Address",
	"ld %sx, ASX = Load S",
	"ldu %sx, ASX = Load S Upper",
	"ldl[.ex] %sx, ASX = Load S Lower",
	"ld2b[.ex] %sx, ASX = Load 2B",
	"ld1b[.ex] %sx, ASX = Load 1B",
	"st %sx, ASX = Store S",
	"stu %sx, ASX = Store S Upper",
	"stl %sx, ASX = Store S Lower",
	"st2b %sx, ASX = Store 2B",
	"st1b %sx, ASX = Store 1B",
	"dld %sx, ASX = Dismissable Load S",
	"dldu %sx, ASX = Dismissable Load Upper",
	"dldl[.ex] %sx, ASX = Dismissable Load Lower",
	"pfch ASX = Pre Fetch",
	"cmov.df[.cf ] %sx, {%sz|M}, {%sy|I} = Conditional Move",
	"addu.l %sx,{%sy|I}, {%sz|M} = Add",
	"addu.w %sx,{%sy|I}, {%sz|M} = Add",
	"adds.w[.ex] %sx, {%sy|I}, {%sz|M} = Add Single",
	"adds. l %sx, {%sy|I}, {%sz|M} = Add",
	"subu.l %sx, {%sy|I}, {%sz|M} = Subtract",
	"subu.w %sx, {%sy|I}, {%sz|M} = Subtract",
	"subs.w[.ex] %sx, {%sy|I}, {%sz|M} = Subtract Single",
	"subs.l %sx, {%sy | I}, {%sz | M} = Subtract",
	"mulu.l %sx, {%sy|I}, {%sz|M} = Multiply",
	"mulu.w %sx, {%sy|I}, {%sz|M} = Multiply",
	"muls.w[.ex] %sx, {%sy|I}, {%sz|M} = Multiply Single",
	"muls.l %sx, {%sy|I}, {%sz|M} = Multiply",
	"muls.l.w %sx, {%sy|I}, {%sz|M} = Multiply",
	"divu.l %sx, {%sy|I}, {%sz|M} = Divide",
	"divu.w %sx, {%sy|I}, {%sz|M} = Divide",
	"divs.w[.ex] %sx, {%sy|I}, {%sz|M} = Divide Single",
	"divs.l %sx, {%sy|I}, {%sz|M} = Divide",
	"cmpu.l %sx, {%sy|I}, {%sz|M} = Compare",
	"cmpu.w %sx, {%sy|I}, {%sz|M} = Compare",
	"cmps.w[.ex] %sx, {%sy|I}, {%sz|M} = Compare Single",
	"cmps.l %sx, {%sy|I}, {%sz|M} = Compare",
	"maxs.w[.ex] %sx, {%sy|I}, {%sz|M} = Compare and Select",
	"mins.w[.ex] %sx, {%sy|I}, {%sz|M} = Maximum/Mininum Single",
	"maxs.l %sx, {%sy|I}, {%sz|M} = Compare and Select",
	"mins.l %sx, {%sy|I}, {%sz|M} = Maximum/Minimum",
	"and %sx, {%sy|I}, {%sz|M} = AND (logical)",
	"or %sx, {%sy|I}, {%sz|M} = OR (logical)",
	"xor %sx, {%sy|I}, {%sz|M} = Exclusive OR (logical)",
	"eqv %sx, {%sy|I}, {%sz|M} = Equivalence (logical)",
	"nnd %sx, {%sy|I}, {%sz|M} = Negate AND (logical)",
	"mrg %sx, {%sy|I}, {%sz|M} = Merge (logical)",
	"pcnt %sx, {%sz|M} = Population Count ",
	"brv %sx, {%sz|M} = Bit Reverse",
	"sll %sx, {%sz|M}, {%sy|N} = Shift Left Logical",
	"sld %sx, {%sz|M}, {%sy|N} = Shift Left Double",
	"srl %sx, {%sz|M}, {%sy|N} = Shift Right Logical",
	"srd %sx, {%sz|M}, {%sy|N} = Shift Right Double",
	"sla.w[.ex] %sx, {%sz|M}, {%sy|N} = Shift Left Arithmetic",
	"sla.l %sx, {%sz|M}, {%sy|N} = Shift Left Arithmetic",
	"sra.w[.ex] %sx, {%sz|M}, {%sy|N} = Shift Right Arithmetic",
	"sra.l %sx, {%sz|M}, {%sy|N} = Shift Right Arithmetic",
	"fadd.d %sx, {%sy|I}, {%sz|M} = Floating Add",
	"fsub.d %sx, {%sy|I}, {%sz|M} = Floating Subtract",
	"fsub.s %sx, {%sy|I}, {%sz|M} = Floating Multiply",
	"fmul.d %sx, {%sy|I}, {%sz|M} = Floating Multiply",
	"fmul.s %sx, {%sy|I}, {%sz|M} = Floating Multiply",
	"fdiv.d %sx, {%sy|I}, {%sz|M} = Floating Divide",
	"fdiv.s %sx, {%sy|I}, {%sz|M} = Floating Divide",
	"fcmp.d %sx, {%sy|I}, {%sz|M} = Floating Compare",
	"fcmp.s %sx, {%sy|I}, {%sz|M} = Floating Compare",
	"fmax.d %sx, {%sy|I}, {%sz|M} = Maximum/Minimum",
	"fmax.s %sx, {%sy|I}, {%sz|M} = Maximum/Minimum",
	"fmin.d %sx, {%sy|I}, {%sz|M} = Maximum/Minimum",
	"fmin.s %sx, {%sy|I}, {%sz|M} = Maximum/Minimum",
	"fadd.q %sx, {%sy|I}, {%sz|M} = Floating Add Quadruple",
	"fsub.q %sx, {%sy|I}, {%sz|M} = Floating Subtract Quadruple",
	"fmul.q %sx, {%sy|I}, {%sz|M} = Floating Multiply Quadruple",
	"fcmp.q %sx, {%sy|I}, {%sz|M} = Floating Compare Quadruple",
	"cvt.w.d[.ex][.rd] %sx, {%sy|I} = Convert to Fixed Point",
	"cvt.w.s[.ex][.rd] %sx, {%sy|I} = Convert to Fixed Point",
	"cvt.l.d[.rd] %sx, {%sy|I} = Convert to Fixed Point",
	"cvt.d.w %sx, {%sy|I} = Convert to Floating Point",
	"cvt.s.w %sx, {%sy|I} = Convert to Floating Point",
	"cvt.d.l %sx, {%sy|I} = Convert to Floating Point",
	"cvt.d.s %sx, {%sy|I} = Convert to Double-format",
	"cvt.d.q %sx, {%sy|I} = Convert to Double-format",
	"cvt.s.d %sx, {%sy|I} = Convert to Single-format",
	"cvt.s.q %sx, {%sy|I} = Convert to Single-format",
	"cvt.q.d %sx, {%sy|I} = Convert to Quadruple-format",
	"cvt.q.s %sx, {%sy|I} = Convert to Quadruple-format",
	"bsic %sx, ASX = Branch and Save IC",
	"vld[.nc] {%vx | %vix}, {%sy | I}, {%sz | Z} = Vector Load",
	"vldu[.nc] {%vx | %vix}, {%sy | I}, {%sz | Z} = Vector Load Upper",
	"vldl[.ex][.nc] {%vx | %vix}, {%sy | I}, {%sz | Z} = Vector Load Lower",
	"vld2d[.nc] {%vx | %vix}, {%sy | I}, {%sz | Z} = Vector Load 2D",
	"vldu2d[.nc] {%vx | %vix}, {%sy | I}, {%sz | Z} = Vector Load Upper 2D",
	"vldl2d[.ex][.nc] {%sy | I}, {%sz | Z}, [, %vm] = Vector Load Lower 2D",
	"vst[.nc][.ot] {%vx | %vix}, {%sy | I}, {%sz | Z} [, %vm] = Vector Store",
	"vstu[.nc][.ot] {%vx | %vix}, {%sy | I}, {%sz | Z} [, %vm] = Vector Store Upper",
	"vstl[.nc][.ot] {%vx | %vix}, {%sy | I}, {%sz | Z} [, %vm] = Vector Store Lower",
	"vst2d[.nc][.ot] | I}, {%sz | Z} [, %vm] = Vector Store 2D",
	"vstu2d[.nc][.ot] {%vx | %vix}, {%sy | I}, {%sz | Z} [, %vm] = Vector Store Upper 2D",
	"vstl2d[.nc][.ot] {%vx | %vix}, {%sy | I}, {%sz | Z} [, %vm] = Vector Store Lower 2D",
	"pfchv[.nc] {%sy | I}, {%sz | Z} = Pre Fetch Vector",
	"lsv {%vx | %vix}({%sy |N}), {%sz | M} = Load S to V",
	"lvs %sx, {%vx | %vix}({%sy | N}) = Load V to S",
	"lvm %vmx, {%sy | N}, {%sz | M} = Load VM",
	"svm %sx, %vmz, {%sy | N} = Save VM N = 0 - 3 ",
	"vbrd  {%vx | %vix}, {%sy | I} [, %vm] = Vector Broadcast",
	"vbrdl {%vx | %vix}, {%sy | I} [, %vm] = Vector Broadcast",
	"vbrdu {%vx | %vix}, {%sy | I} [, %vm] = Vector Broadcast",
	"pvbrd {%vx | %vix}, {%sy | I} [, %vm] = Vector Broadcast",
	"vmv {%vx | %vix}, {%sy | N}, {%vz | %vix} [, %vm] = Vector Move",
	"vaddu.df {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Add",
	"pvaddu.lo {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Add",
	"pvaddu.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Add",
	"pvaddu {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Add",
	"vadds.w[.ex] {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Add Single",
	"pvadds.lo[.ex] {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Add Single",
	"pvadds.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Add Single",
	"pvadds {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Add Single",
	"vadds.l {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Add",
	"vsubu.df {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Subtract",
	"vsubu.lo {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Subtract",
	"pvsubu.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Subtract",
	"pvsubu {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Subtract",
	"vsubs.w[.ex] {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Subtract Single",
	"pvsubs.lo[.ex] {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Subtract Single",
	"pvsubs.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Subtract Single",
	"pvsubs {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Subtract Single",
	"vsubs.l {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Subtract",
	"vmulu.df {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Multiply",
	"vmuls.w[.ex] {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Multiply Single",
	"vmuls.l {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Multiply",
	"vmuls.l.w {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [,%vm] = Vector Multiply",
	"vdivu.df {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix | %sy | I} [, %vm] = Vector Divide",
	"vdivs.w[.ex] {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix | %sy | I} [,%vm] = Vector Divide Single",
	"vdivs.l {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix | %sy | I} [, %vm] = Vector Divide",
	"vcmpu.df {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare",
	"pvcmpu.lo {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare",
	"pvcmpu.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare",
	"pvcmpu {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare",
	"vcmps.w[.ex] {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare Single",
	"pvcmps.lo[.ex] {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare Single",
	"pvcmps.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare Single",
	"pvcmps {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare Single",
	"vcmps.l {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare",
	"vmaxs.w[.ex] {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare and Select Maximum/Minimum Single",
	"pvmaxs.lo[.ex] {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare and Select Maximum/Minimum Single",
	"pvmaxs.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare and Select Maximum/Minimum Single",
	"pvmaxs {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare and Select Maximum/Minimum Single",
	"vmins.w[.ex] {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare and Select Maximum/Minimum Single",
	"pvmins.lo[.ex] {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare and Select Maximum/Minimum Single",
	"pvmins.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare and Select Maximum/Minimum Single",
	"pvmins {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare and Select Maximum/Minimum Single",
	"vmaxs.l {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare and Select Maximum/Minimum",
	"vmins.l {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Compare and Select Maximum/Minimum",
	"vand {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector AND",
	"pvand.lo {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector AND",
	"pvand.up {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector AND",
	"pvand {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector AND",
	"vor {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector OR",
	"pvor.lo {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector OR",
	"pvor.up {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector OR",
	"pvor {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector OR",
	"vxor {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector Exclusive OR",
	"pvxor.lo {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector Exclusive OR",
	"pvxor.up {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector Exclusive OR",
	"pvxor {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector Exclusive OR",
	"veqv {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector Equivalence",
	"pveqv.lo {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector Equivalence",
	"pveqv.up {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector Equivalence",
	"pveqv {%vx | %vix}, {%vy | %vix | %sy | M}, {%vz | %vix} [, %vm] = Vector Equivalence",
	"vldz {%vx | %vix}, {%vz | %vix} [, %vm] = Vector Leading Zero Count",
	"pvldz.lo {%vx | %vix}, {%vz | %vix} [, %vm] = Vector Leading Zero Count",
	"pvldz.up {%vx | %vix}, {%vz | %vix} [, %vm] = Vector Leading Zero Count",
	"pvldz {%vx | %vix}, {%vz | %vix} [, %vm] = Vector Leading Zero Count",
	"vpcnt {%vx | %vix}, {%vz | %vix} [, %vm] = Vector Population Count",
	"pvpcnt.lo {%vx | %vix}, {%vz | %vix} [, %vm] = Vector Population Count",
	"pvpcnt.up {%vx | %vix}, {%vz | %vix} [, %vm] = Vector Population Count",
	"pvpcnt {%vx | %vix}, {%vz | %vix} [, %vm] = Vector Population Count",
	"vbrv {%vx | %vix}, {%vz | %vix} [, %vm] = Vector Bit Reverse",
	"pvbrv.lo {%vx | %vix}, {%vz | %vix} [, %vm] = Vector Bit Reverse",
	"pvbrv.up {%vx | %vix}, {%vz | %vix} [, %vm] = Vector Bit Reverse",
	"pvbrv {%vx | %vix}, {%vz | %vix} [, %vm] = Vector Bit Reverse",
	"vseq {%vx | %vix} [, %vm] = Vector Sequential Number",
	"pvseq.lo {%vx | %vix} [, %vm] = Vector Sequential Number",
	"pvseq.up {%vx | %vix} [, %vm] = Vector Sequential Number",
	"pvseq {%vx | %vix} [, %vm] = Vector Sequential Number",
	"vsll {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy | N} [, %vm] = Vector Shift Left Logical",
	"pvsll.lo {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy | N} [, %vm] = Vector Shift Left Logical",
	"pvsll.up {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy} [, %vm] = Vector Shift Left Logical",
	"pvsll {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy} [, %vm] = Vector Shift Left Logical",
	"vsld {%vx | %vix}, ({%vy | %vix}, {%vz | %vix}), {%sy | N} [, %vm] = Vector Shift Left Double",
	"vsrl {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy | N} [, %vm] = Vector Shift Right Logical",
	"pvsrl.lo {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy | N} [, %vm] = Vector Shift Right Logical",
	"pvsrl.up {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy} [, %vm] = Vector Shift Right Logical",
	"pvsrl {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy} [, %vm] = Vector Shift Right Logical",
	"vsrd {%vx | %vix}, ({%vy | %vix}, {%vz | %vix}), {%sy | N} [, %vm] = Vector Shift Right Double",
	"vsla.w[.ex] {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy | N} [, %vm] = Vector Shift Left Arighmetic",
	"pvsla.lo[.ex] {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy | N} [, %vm] = Vector Shift Left Arighmetic",
	"pvsla.up {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy} [, %vm] = Vector Shift Left Arighmetic",
	"pvsla {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy} [, %vm] = Vector Shift Left Arighmetic",
	"vsla.l {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy | N} [, %vm] = Vector Shift Left Arithmetic",
	"vsra.w[.ex] {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy | N} [, %vm] = Vector Shift Right Arithmetic",
	"pvsra.lo[.ex] {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy | N} [, %vm] = Vector Shift Right Arithmetic",
	"pvsra.up {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy} [, %vm] = Vector Shift Right Arithmetic",
	"pvsra {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy} [, %vm] = Vector Shift Right Arithmetic",
	"vsra.l {%vx | %vix}, {%vz | %vix}, {%vy | %vix | %sy | N} [, %vm] = Vector Shift Right Arithmetic",
	"vsfa {%vx | %vix}, {%vz | %vix}, {%sy | N}, {%sz | M} [, %vm] = Vector Shift Left and Add",
	"vfadd.df {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Add",
	"pvfadd.lo {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Add",
	"pvfadd.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Add",
	"pvfadd {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Add",
	"vfsub.df {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Subtract",
	"pvfsub.lo {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Subtract",
	"pvfsub.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Subtract",
	"pvfsub {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Subtract",
	"vfmul.df {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Multiply",
	"pvfmul.lo {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Multiply",
	"pvfmul.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Multiply",
	"pvfmul {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Multiply",
	"vfdiv.df {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix {%vx | %vix}, {%vy | %sy | I} [, %vm] = Vector Floating Divide",
	"vfsqrt.df {%vx | %vix}, {%vy | %vix} [, %vm]  = Vector Floating Square Root",
	"vfcmp.df | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Compare",
	"pvfcmp.lo {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Compare",
	"pvfcmp.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Compare",
	"vfcmp {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Compare and Select Maximum/Minimum",
	"vfmax.df | %vix | %sy | I}, {%vz | %vix} {%vx | %vix}, {%vy [, %vm] = Vector Floating Compare and Select Maximum/Minimum",
	"pvfmax.lo {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Compare and Select Maximum/Minimum",
	"pvfmax.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Compare and Select Maximum/Minimum",
	"pvfmax {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Compare and Select Maximum/Minimum",
	"vfmin.df {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Compare and Select Maximum/Minimum",
	"pvfmin.lo {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Compare and Select Maximum/Minimum",
	"pvfmin.up {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Compare and Select Maximum/Minimum",
	"pvfmin {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Floating Compare and Select Maximum/Minimum",
	"vfmad.df {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} }| { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Multiply Add",
	"pvfmad.lo {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Multiply Add",
	"pvfmad.up {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix} , {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Multiply Add",
	"pvfmad {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } |{ {%sy | I}, {%vz | %vix} }| { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Multiply Add",
	"vfmsb.df {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Multiply Subtract",
	"pvfmsb.lo {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Multiply Subtract",
	"pvfmsb.up {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Multiply Subtract",
	"pvfmsb {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Multiply Subtract",
	"vfnmad.df {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Negative Multiply Add",
	"pvfnmad.lo {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Negative Multiply Add",
	"pvfnmad.up {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Negative Multiply Add",
	"pvfnmad {%vx | %vix}, { { {%vy | %vix} | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} }, {%vw | %vix} [, %vm] = Vector Floating Fused Negative Multiply Add",
	"vfnmsb.df {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Negative Multiply Subtract",
	"pvfnmsb.lo {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Negative Multiply Subtract",
	"pvfnmsb.up {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Negative Multiply Subtract",
	"pvfnmsb {%vx | %vix}, { { {%vy | %vix}, {%vz | %vix} } | { {%sy | I}, {%vz | %vix} } | { {%vy | %vix}, {%sy | I} } }, {%vw | %vix} [, %vm] = Vector Floating Fused Negative Multiply Subtract",
	"vrcp.df {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Floating Reciprocal",
	"pvrcp.lo {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Floating Reciprocal",
	"pvrcp.up {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Floating Reciprocal",
	"pvrcp {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Floating Reciprocal",
	"vrsqrt.df[.nex] {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Floating Reciprocal",
	"pvrsqrt.lo[.nex] {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Floating Reciprocal",
	"pvrsqrt.up[.nex] {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Floating Reciprocal",
	"pvrsqrt[.nex] {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Floating Reciprocal",
	"vcvt.w.df[.ex][.rd] {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Convert to Fixed Point",
	"pvcvt.w.s.lo[.rd] {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Convert to Fixed Point",
	"pvcvt.w.s.up[.rd] {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Convert to Fixed Point",
	"pvcvt.w.s[.rd] {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Convert to Fixed Point",
	"vcvt.l.d[.rd] {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Convert to Fixed Point",
	"vcvt.df.w {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Convert to Floating Point",
	"pvcvt.s.w.lo {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Convert to Floating Point",
	"pvcvt.s.w.up {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Convert to Floating Point",
	"pvcvt.s.w {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Convert to Floating Point",
	"vcvt.d.l {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Convert to Floating",
	"vcvt.d.s {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Convert to Single-Format",
	"vcvt.s.d {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Convert to Double-Format",
	"vmrg {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Merge",
	"vmrg.l {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Merge",
	"vmrg.w {%vx | %vix}, {%vy | %vix | %sy | I}, {%vz | %vix} [, %vm] = Vector Merge",
	"vshf {%vx | %vix}, {%vy | %vix}, {%vz | %vix}, {%sy | N} = Vector Shuffle N = 0 - 15",
	"vcp {%vx | %vix}, {%vz | %vix}[, %vm] =  Vector Compress",
	"vex {%vx | %vix}, {%vz | %vix}[, %vm] = Vector Expand",
	"vfmk.l.cf %vmx, {%vz | %vix} [, %vm] = Vector Form Mask If cf is \"af\" or \"at\", %vz can be omitted.  If cf is \"at\", cf can be omitted.",
	"vfmk.w.cf %vmx, {%vz | %vix} [, %vm] = Vector Form Mask Single If cf is \"af\" or \"at\", %vz can be If cf is \"at\", cf can be omitted.  omitted.",
	"pvfmk.w.lo.cf %vmx, {%vz | %vix} [, %vm] = Vector Form Mask Single If cf is \"af\" or \"at\", %vz can be If cf is \"at\", cf can be omitted.  omitted.",
	"pvfmk.w.up.cf %vmx, {%vz | %vix} [, %vm] = Vector Form Mask Single If cf is \"af\" or \"at\", %vz can be If cf is \"at\", cf can be omitted.  omitted.",
	"vfmk.df.cf %vmx, {%vz | %vix} [, %vm] = Vector Form Mask Floating Point If cf is \"af\" or \"at\", %vz can be omitted.  If cf is \"at\", cf can be omitted.",
	"pvfmk.s.lo.cf %vmx, {%vz | %vix} [, %vm] = Vector Form Mask Floating Point If cf is \"af\" or \"at\", %vz can be omitted.  If cf is \"at\", cf can be omitted.",
	"pvfmk.s.up.cf %vmx, {%vz | %vix} [, %vm] = Vector Form Mask Floating Point If cf is \"af\" or \"at\", %vz can be omitted.  If cf is \"at\", cf can be omitted.",
	"vsum.w[.ex] {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Sum Single",
	"vsum.l {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Sum",
	"vfsum.df {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Floating Sum",
	"vrmaxs.w.pos[.ex] {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Maximum/Minimum Single",
	"vrmins.w.pos[.ex] {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Maximum/Minimum Single",
	"vrmaxs.l.pos {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Maximum/Minimum Single",
	"vrmins.l.pos {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Maximum/Minimum Single",
	"vfrmax.df.pos {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Floating Maximum/Minimum",
	"vfrmin.df.pos {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Floating Maximum/Minimum",
	"vrand {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Reduction AND",
	"vror {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Reduction OR",
	"vrxor {%vx | %vix}, {%vy | %vix} [, %vm] = Vector Reduction Exclusive OR",
	"vfia.df  {%vx | %vix}, {%vy | %vix}, {%sy | I} = Vector Floating Iteration Add ",
	"vfis.df {%vx | %vix}, {%vy | %vix}, {%sy | I} = Vector Floating Iteration Subtract",
	"vfim.df {%vx | %vix}, {%vy | %vix}, {%sy | I} = Vector Floating Iteration Multiply",
	"vfiam.df {%vx | %vix}, {%vy | %vix}, {%vz | %vix}, {%sy | I} = Vector Floating Iteration Add and Multiply",
	"vfism.df {%vx | %vix}, {%vy | %vix}, {%vz | %vix}, {%sy | I} = Vector Floating Iteration Subtract and Multiply",
	"vfima.df {%vx | %vix}, {%vy | %vix}, {%vz | %vix}, {%sy | I} = Vector Floating Iteration Multiply and Add",
	"vfims.df {%vx | %vix}, {%vy | %vix}, {%vz | %vix}, {%sy | I} = Vector Floating Iteration Multiply and Subtract",
	"vgt[.{nc}] {%vx | %vix}, {%vy | %vix | %sw}, {%sy | I}, {%sz | Z} [,%vm] = Vector Gather",
	"vgtu[.{nc}] {%vx | %vix}, {%vy | %vix | %sw}, {%sy | I}, {%sz | Z} [,%vm] = Vector Gather Upper",
	"vgtl[.ex][.{nc}] {%vx | %vix}, {%vy | %vix | %sw}, {%sy | I}, {%sz | Z} [,%vm] = Vector Gather Lower",
	"vsc[.nc][.ot] {%vx | %vix}, {%vy | %vix | %sw}, {%sy | I}, {%sz | Z} [, %vm] = Vector Scatter",
	"vscu[.nc][.ot] {%vx | %vix}, {%vy | %vix | %sw}, {%sy | I}, {%sz | Z} [, %vm] = Vector Scatter Upper",
	"vscl[.nc][.ot] {%vx | %vix}, {%vy | %vix | %sw}, {%sy | I}, {%sz | Z} [, %vm] = Vector Scatter Lower",
	"andm %vmx, %vmy, %vmz = AND VM",
	"orm %vmx, %vmy, %vmz = OR VM",
	"xorm %vmx, %vmy, %vmz = Exclusive OR VM",
	"eqvm %vmx, %vmy, %vmz = Equivalence VM",
	"nndm %vmx, %vmy, %vmz = Negate AND VM ",
	"negm %vmx, %vmy = Negate VM",
	"pcvm %sx, %vmy = Population Count of VM",
	"lzvm %sx, %vmy = Leading Zero of VM",
	"tovm %sx, %vmy = Trailing One of VM",
	"lvl {%sy | I} = Load VL",
	"svl %sx = Save VL",
	"smvl %sx = Save Maximum Vector",
	"lvix {%sy | N} = Load Vector Data Index N = 0 - 63",
	"sic %sx = Save Instruction Counter",
	"lpm %sy = Load Program Mode Flags",
	"spm %sx = Save Program Mode Flags",
	"lfr {%sy | N} = Load Flag Register N = 0 - 63",
	"sfr %sx = Safe Flag Register",
	"smir %sx, I = Save Miscellaneous Register",
	"smir %sx, %usrcc = Save Miscellaneous Register",
	"smir %sx, %psw = Save Miscellaneous Register I = 0 - 2, 7 - 11, 16 - 30",
	"smir %sx, %sar = Save Miscellaneous Register",
	"smir %sx, %pmmr = Save Miscellaneous Register MM = 0 - 3",
	"smir %sx, %pmcrNN = Save Miscellaneous Register NN = 0 - 14 ",
	"smir %sx, %pmcNN = Save Miscellaneous Register",
	"nop = No Operation",
	"monc [N, N, N] = Monitor Call",
	"monc.hdb [N, N, N] = Monitor Call 2nd and 3rd operand : N = 0 - 255",
	"lcr %sx, {%sy|I}, {%sz|Z} = Load Communication Register",
	"scr %sx, {%sy|I}, {%sz|Z} = Store Communication Register",
	"tscr %sx, {%sy|I}, {%sz|Z} = Test and Set Communication Register",
	"fidcr %sx, {%sy|I}, I = Fetch and Increment/Decrement CR 3rd operand : N = 0 7",
	"ts1am.l %sx, {%sz | AS}, {%sy|N} = Test and Set 1 AM",
	"ts1am.w %sx, {%sz | AS}, {%sy|N} = Test and Set 1 AM",
	"ts2am %sx, {%sz | AS}, {%sy|N} = Test and Set 2 AM",
	"ts3am %sx, {%sz | AS}, {%sy|N} = Test and Set 3 AM N = 0 or 1",
	"atmam %sx, {%sz | AS}, {%sy|N} = Atomic AM N = 0 - 2",
	"cas.l %sx, {%sz | AS}, {%sy|I} = Compare and Swap",
	"cas.w %sx, {%sz | AS}, {%sy|I} = Compare and Swap",
	"fencei = fence",
	"fencem I = fence I = 1 - 3",
	"fencec I = fence I = 1 - 7",
	"svob = Set Vector Out-of-order memory access Boundary",
	"bswp %sx, {%sz | M}, I = Byte Swap I = 0 or 1",
	"lhm.b %sx, HM = Load Host Memory",
	"lhm.h %sx, HM = Load Host Memory",
	"lhm.w %sx, HM = Load Host Memory",
	"lhm.l %sx, HM = Load Host Memory",
	"shm.b %sx, HM = Store Host Memory",
	"shm.h %sx, HM = Store Host Memory",
	"shm.w %sx, HM = Store Host Memory",
	"shm.l %sx, HM = Store Host Memory"}
