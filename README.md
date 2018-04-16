# veass

VE assistant

Allows to browse assembler code with annotations and compare
with source code.

## usage

compile with `ncc/nc++/nfort -g -S`

and view with

    veass [-s srcdir[,srcdir]] ass.s

if source is not in same directory as assemblerfile, a list of search directories
can be specified.

## keys

After loading the assembler file, the file can be navigated with cursor keys,
`pageup`, `pagedown` and `pos1` and `end` keys. A short help is displayed on start and when
pressing `h`. You can exit any time with `q`.

The top panel shows the assember file, current line is drawn in bold mode.
if `return` is pressed, the lower panel shows an explanation of the assemblerfile
instruction of the current line.

with `v`, a second panel with the source file (if found) can be opened, keyboard
focus can be changed with `TAB`. The source view can be closed with `V`.
Only if the source view is opened, source can be displayed when pressing `return`.

If `return` is pressed in source panel, the according assembler lines are highlighted
and the first marked line is shown.

if `return` is pressed in assembler panel, the according source line is highlighted
and shown in the source view, and the instruction is explained (for VE and x86).

The marking of lines in both views can be cleared with `c`.

`/` starts a search, `n` and `p` jump to next or previous search hit, marked region or global label.
