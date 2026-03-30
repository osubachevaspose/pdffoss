create Golang library source code to parse PDF file with classic cross-reference table
separate Golang struct must be created for each basic PDF object: null, boolean, numeric, name, string, array, dictionary, stream
do not implement graphics, encryption processing
do not add PDF generating but only parsing of PDF file

================================================================

reread pdfhelper repository and add unit test for findStartXRef function

================================================================

add unit test to parser_test.go that tests parsing of external PDF file with classic cross-reference table

================================================================

add standard font Helvetica to generated pdfparser\testdata\classic_xref.pdf file
and use this font to render text in this file
