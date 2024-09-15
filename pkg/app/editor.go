package app

import "github.com/heartandu/easyrpc/pkg/editor"

func (a *App) editor() editor.Editor {
	cmd := "nano"
	if e := a.cfg.Editor.Cmd; e != "" {
		cmd = e
	}

	return editor.NewCmdEditor(a.fs, cmd)
}
