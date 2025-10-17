package pdfs

type PaperSize struct {
	Name   string
	Width  float64 // in `pt` (1" = 72pts)
	Height float64 // in `pt`
}

var (
	LetterSize = PaperSize{Name: "Letter", Width: 612, Height: 792}         // 8.5" x 11"
	A4Size     = PaperSize{Name: "A4", Width: 595.27559, Height: 595.27559} // 210mm x 297mm
)
