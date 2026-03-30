package pdfparser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func buildSimplePDF() []byte {
	obj1 := "1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n"
	obj2 := "2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n"
	obj3 := "3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 300 144] /Contents 4 0 R >>\nendobj\n"
	stream := "BT\n/F1 12 Tf\n72 72 Td\n(Hello) Tj\nET\n"
	obj4 := fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n%sendstream\nendobj\n", len(stream), stream)

	var b strings.Builder
	b.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	for _, obj := range []string{obj1, obj2, obj3, obj4} {
		offsets = append(offsets, b.Len())
		b.WriteString(obj)
	}
	xrefOffset := b.Len()
	b.WriteString("xref\n")
	b.WriteString("0 5\n")
	b.WriteString("0000000000 65535 f \n")
	for i := 1; i <= 4; i++ {
		b.WriteString(fmt.Sprintf("%010d 00000 n \n", offsets[i]))
	}
	b.WriteString("trailer\n")
	b.WriteString("<< /Size 5 /Root 1 0 R >>\n")
	b.WriteString("startxref\n")
	b.WriteString(fmt.Sprintf("%d\n", xrefOffset))
	b.WriteString("%%EOF\n")
	return []byte(b.String())
}

func TestParseClassicXRefPDF(t *testing.T) {
	doc, err := Parse(buildSimplePDF())
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if doc.Version != "1.4" {
		t.Fatalf("unexpected version %q", doc.Version)
	}

	if len(doc.Objects) != 4 {
		t.Fatalf("expected 4 objects, got %d", len(doc.Objects))
	}

	obj4, ok := doc.Objects[ObjectID{Number: 4, Generation: 0}]
	if !ok {
		t.Fatalf("missing object 4 0")
	}
	stream, ok := obj4.Value.(PDFStream)
	if !ok {
		t.Fatalf("object 4 is not a stream")
	}
	if !strings.Contains(string(stream.Data), "Hello") {
		t.Fatalf("unexpected stream data: %q", string(stream.Data))
	}
}

func TestParseClassicXRefExternalPDF(t *testing.T) {
	pdfPath := filepath.Join("testdata", "classic_xref.pdf")
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) failed: %v", pdfPath, err)
	}

	doc, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if doc.Version != "1.4" {
		t.Fatalf("unexpected version %q", doc.Version)
	}

	if len(doc.Objects) != 5 {
		t.Fatalf("expected 5 objects, got %d", len(doc.Objects))
	}

	root, ok := doc.Trailer.Entries["Root"]
	if !ok {
		t.Fatal("missing trailer Root entry")
	}
	rootRef, ok := root.(PDFIndirectRef)
	if !ok {
		t.Fatalf("trailer Root has unexpected type %T", root)
	}
	if rootRef.ObjectNumber != 1 || rootRef.GenerationNumber != 0 {
		t.Fatalf("unexpected trailer Root reference: %+v", rootRef)
	}

	page, ok := doc.Objects[ObjectID{Number: 3, Generation: 0}]
	if !ok {
		t.Fatalf("missing object 3 0")
	}
	pageDict, ok := page.Value.(PDFDictionary)
	if !ok {
		t.Fatalf("object 3 is not a dictionary")
	}
	resourcesObj, ok := pageDict.Get("Resources")
	if !ok {
		t.Fatal("page is missing Resources")
	}
	resources, ok := resourcesObj.(PDFDictionary)
	if !ok {
		t.Fatalf("page Resources has unexpected type %T", resourcesObj)
	}
	fontsObj, ok := resources.Get("Font")
	if !ok {
		t.Fatal("page Resources is missing Font")
	}
	fonts, ok := fontsObj.(PDFDictionary)
	if !ok {
		t.Fatalf("page Font resources has unexpected type %T", fontsObj)
	}
	fontRefObj, ok := fonts.Get("F1")
	if !ok {
		t.Fatal("page Font resources is missing F1")
	}
	fontRef, ok := fontRefObj.(PDFIndirectRef)
	if !ok {
		t.Fatalf("page font F1 has unexpected type %T", fontRefObj)
	}
	if fontRef.ObjectNumber != 5 || fontRef.GenerationNumber != 0 {
		t.Fatalf("unexpected F1 font reference: %+v", fontRef)
	}

	obj4, ok := doc.Objects[ObjectID{Number: 4, Generation: 0}]
	if !ok {
		t.Fatalf("missing object 4 0")
	}
	stream, ok := obj4.Value.(PDFStream)
	if !ok {
		t.Fatalf("object 4 is not a stream")
	}
	if string(stream.Data) != "BT\n/F1 12 Tf\n72 72 Td\n(Hello) Tj\nET\n" {
		t.Fatalf("unexpected stream data: %q", string(stream.Data))
	}

	font, ok := doc.Objects[ObjectID{Number: 5, Generation: 0}]
	if !ok {
		t.Fatalf("missing object 5 0")
	}
	fontDict, ok := font.Value.(PDFDictionary)
	if !ok {
		t.Fatalf("object 5 is not a dictionary")
	}
	baseFontObj, ok := fontDict.Get("BaseFont")
	if !ok {
		t.Fatal("font is missing BaseFont")
	}
	baseFont, ok := baseFontObj.(PDFName)
	if !ok {
		t.Fatalf("font BaseFont has unexpected type %T", baseFontObj)
	}
	if baseFont.Value != "Helvetica" {
		t.Fatalf("unexpected BaseFont %q", baseFont.Value)
	}
}

func TestParseBasicObjects(t *testing.T) {
	p := &reader{data: []byte("[null true false 12 -3.5 /A (B) <4344> << /K 1 >>]")}
	obj, err := p.parseObject()
	if err != nil {
		t.Fatalf("parseObject failed: %v", err)
	}
	arr, ok := obj.(PDFArray)
	if !ok {
		t.Fatalf("expected array, got %T", obj)
	}
	if len(arr.Elements) != 9 {
		t.Fatalf("unexpected element count %d", len(arr.Elements))
	}
}

func TestFindStartXRef(t *testing.T) {
	pdfData := []byte("%PDF-1.4\n1 0 obj\n<<>>\nendobj\nstartxref\n123\n%%EOF\n")
	offset, err := findStartXRef(pdfData)
	if err != nil {
		t.Fatalf("findStartXRef returned error: %v", err)
	}
	if offset != 123 {
		t.Fatalf("expected startxref offset 123, got %d", offset)
	}
}

func TestFindStartXRefMissing(t *testing.T) {
	_, err := findStartXRef([]byte("%PDF-1.4\n%%EOF\n"))
	if err == nil {
		t.Fatal("expected error for missing startxref")
	}
}
