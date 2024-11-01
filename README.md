# orca-st (Go Edition)
(Hassle-free) extraction of state informations from [ORCA](https://orcaforum.kofo.mpg.de) 
output files. Threshold based printing is possible.

Don't like Go? Try the [Python Edition](https://github.com/radi0sus/orca_st)!

## Quick start
 Start the script with:
```console
go run orca-st.go -f filename
```
or build an executable first with:
```console
go build orca-st.go
```
and start with:
```console
orca-st(.exe) -f filename
```
it will show the table in the console. The table will probably exceed the size of
your console window and the table might therefore look unfamiliar.

Start the script with:
```console
go run orca-st.go -f filename > filename.md
```
or
```console
orca-st(.exe) -f filename > filename.md
```
it will save the table in markdown format.

Convert markdown to docx (install [PANDOC](https://pandoc.org) first):
```console
pandoc filename.md -o filename.docx
```
This will convert the markdown file to a docx file. Open it with your favorite
word processor. Convert the file to even more formats such as HTML, PDF or TeX with PANDOC.

## Command-line options
- `-f filename`, required: filename
- `-t` `N`, optional: set a threshold in %. Transitions below the threshold value will not be printed (default is `N = 0`)
- `-nto`, optional: process all or selected states for natural transition orbitals (NTO)
- `-tr`, optional: show 'Transition' in case of ORCA 6 output files

## Code options
You can change the table header in the script (take care of the row size if necessary). 

## Remarks
- The data are taken from the section "ABSORPTION SPECTRUM VIA TRANSITION ELECTRIC DIPOLE MOMENTS".
- Only tested with "normal" outputs (including NTO) from TD-DFT calculations.
- Selected transitions that are below the threshold will not be printed in the table. This may result in empty cells.
- If NTO transitions are present int the output file and NTO transitions should be printed, use the `-nto` keyword. 
Otherwise, do not use the `-nto` keyword.
- The script used two unicode characters, namely "⁻¹". Please have a look at the script if you experience any issues. The easiest
solution is to replace "⁻¹" with the ascii characters "-1".

## Examples
See the [Python edition](https://github.com/radi0sus/orca_st).
Differences:
To open a file use `-f filename`.
The `-s` option is not available in the Go edition.
