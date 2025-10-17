package pdfs

import "io"

// Writer — minimal, stream-style, append-only PDF writer. No page navigation
// T: Concrete Template Type -> depends on each implementation
type Writer[T any] interface {
	PaperSize() PaperSize
	Orientation() string

	TemplateStore() *TemplateStore[T]
	ImportPageAsTemplate(filepath string, pageNum int, storeKey string) error

	AddBlankPage()
	AddTemplatePage(storeKey string) bool

	SetFont(family string, style string, size float64)

	Text(x float64, y float64, text string)

	WriteTo(w io.Writer) (int64, error)
	WriteToFile(filepath string) error
	ProduceBytes() ([]byte, error)
}
