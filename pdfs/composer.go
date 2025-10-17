package pdfs

import "io"

// Composer — buffered, reorderable, editable PDF document builder.
type Composer[T any] interface {
	PaperSize() PaperSize
	Orientation() string

	TemplateStore() *TemplateStore[T]
	ImportPageAsTemplate(filepath string, pageNum int, storeKey string) error

	GetCurrentPageIndex() int
	SelectPage(index int) error
	MovePage(from int, to int) error

	AppendBlankPage()
	AppendTemplatePage(storeKey string) error
	InsertPage(index int) int
	InsertTemplatePage(index int, storeKey string) error
	PrependPage() int
	PrependTemplatePage(storeKey string) error

	SetFont(family string, style string, size float64)
	Text(x float64, y float64, text string)

	WriteTo(w io.Writer) (int64, error)
	WriteToFile(filepath string) error
	ProduceBytes() ([]byte, error)
}
