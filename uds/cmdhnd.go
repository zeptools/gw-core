package uds

import (
	"context"
	"io"
)

type CmdHnd struct {
	Desc  string
	Usage string
	Fn    func(ctx context.Context, args []string, out io.Writer)
}
