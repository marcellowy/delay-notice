package main

import (
	_ "delay-notice/internal/packed"

	"github.com/gogf/gf/v2/os/gctx"

	"delay-notice/internal/cmd"
)

func main() {
	cmd.Main.Run(gctx.GetInitCtx())
}
