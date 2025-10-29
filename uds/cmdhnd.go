package uds

import (
	"io"
)

type CmdHnd struct {
	Desc  string
	Usage string
	Fn    func(args []string, w io.Writer) error
}
