package gui

import (
	"github.com/marrow16/gogol/gui/shortcuts"
	"io"
	"log/slog"
	"os"
	"strings"
)

func (c *Core) shortcutLog(msgf string, repeats []int) {
	if (strings.HasPrefix(msgf, `"`) && strings.HasSuffix(msgf, `"`)) || (strings.HasPrefix(msgf, `'`) && strings.HasSuffix(msgf, `'`)) {
		msgf = msgf[1 : len(msgf)-1]
	}
	msg := c.shortcutFormat(msgf, repeats) + "\n"
	_, _ = c.shortcutLogger.Write([]byte(msg))
}

func (c *Core) setupShortcutLogger(intercepts ...io.Writer) *shortcutsLogger {
	if fp, err := resolveSavePath("./output.log"); err == nil {
		var aw io.Writer
		if len(intercepts) > 0 {
			aw = io.MultiWriter(intercepts...)
		}
		return &shortcutsLogger{
			filename:   fp,
			additional: aw,
		}
	}
	return nil
}

func (c *Core) setupShortcutErrorLogger(intercepts ...io.Writer) *shortcutsLogger {
	if fp, err := resolveSavePath("./errors.log"); err == nil {
		el := &shortcutsLogger{
			filename: fp,
		}
		w := io.Writer(el)
		if len(intercepts) > 0 {
			w = io.MultiWriter(append([]io.Writer{w}, intercepts...)...)
		}
		el.logger = slog.New(slog.NewJSONHandler(w, nil))
		return el
	}
	return nil
}

type shortcutsLogger struct {
	filename   string
	additional io.Writer
	logger     *slog.Logger
	callStack  []string
}

func (l *shortcutsLogger) Write(p []byte) (n int, err error) {
	if l != nil {
		if f, fErr := os.OpenFile(l.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); fErr == nil {
			defer func() {
				_ = f.Close()
			}()
			return f.Write(p)
		}
		if l.additional != nil {
			_, _ = l.additional.Write(p)
		}
	}
	return n, err
}

func (l *shortcutsLogger) Error(err error, ah *shortcuts.ActionHolder) {
	if err != nil && l != nil && l.logger != nil {
		args := make([]any, 0)
		if len(l.callStack) > 0 {
			args = append(args, "shortcut", l.callStack[len(l.callStack)-1])
			if len(l.callStack) > 1 {
				args = append(args, "call-stack", strings.Join(l.callStack, " "))
			}
		}
		if ah != nil {
			action := ah.Action.String()
			if ah.IsInc {
				action += "++"
			}
			if ah.IsDec {
				action += "--"
			}
			args = append(args, "action", action)
			if ah.HasOperand {
				args = append(args, "operand", ah.Operand)
			}
		}
		l.logger.Error(err.Error(), args...)
	}
}

func (l *shortcutsLogger) Warn(msg string, ah *shortcuts.ActionHolder, args ...any) {
	if l != nil && l.logger != nil {
		if len(l.callStack) > 0 {
			args = append(args, "shortcut", l.callStack[len(l.callStack)-1])
			if len(l.callStack) > 1 {
				args = append(args, "call-stack", strings.Join(l.callStack, " "))
			}
		}
		if ah != nil {
			action := ah.Action.String()
			if ah.IsInc {
				action += "++"
			}
			if ah.IsDec {
				action += "--"
			}
			args = append(args, "action", action)
			if ah.HasOperand {
				args = append(args, "operand", ah.Operand)
			}
		}
		l.logger.Warn(msg, args...)
	}
}
