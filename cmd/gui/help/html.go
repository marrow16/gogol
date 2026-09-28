package help

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/text"
	"gioui.org/unit"
	"github.com/marrow16/gogol/cmd/gui/help/images"
	"html"
	"image/png"
	"io"
	"slices"
	"strconv"
	"strings"
)

const htmlStyling = `
    <style>
		body {
			font-size: 110%;
		}
		code {
			background-color: rgba(128,128,128,0.25);
		}
		pre {
			background-color: rgba(128,128,128,0.25);
			border: 1px solid black;
			border-radius: 4px;
			padding: 4px;
			margin: 0;
		}
		h1,h2,h3,h4,h5,h6 {
			margin: 0;
		}
		details {
			border: 1px solid black;
			border-radius: 4px;
			padding: 4px;
		}
		.link {
			color: rgb(102,128,230);
		}
		.button {
			padding: 0 4px 0 4px;
			border: 1px solid rgb(128,128,128);
			border-radius: 4px;
			background-color: rgba(128,128,128,0.25);
		}
		.key {
			padding: 0 4px 0 4px;
			border: 1px solid black;
			border-radius: 4px;
		}
		.hanging {
			display: flex;
		}
		table {
			text-align: left;
			border-collapse: collapse;
		}
		table.bordered, table.bordered th, table.bordered td {
			border: 1px solid rgb(128, 128, 128);
		}
		table th, table td {
			padding: var(--cell-padding);
		}
	</style>`

func copyIndexHtml(gtx layout.Context) {
	var hb strings.Builder
	hb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>GoGoL Help</title>` + htmlStyling + `
    <style>
		body {
			margin: 0;
			padding: 0;
		}
		div.help {
			height: 100vh;
			overflow: hidden;
			display: flex;
			flex-direction: column;
		}
		div.title {
			font-size: 1.5em;
			font-weight: bold;
			flex: 0 0 auto;
			border-bottom: 1px solid rgb(128, 128, 128);
			padding: 4px;
			position: relative;
		}
		div.content {
			flex: 1;
			overflow-y: auto;
			min-height: 0;
			padding: 4px;
			display: none;
		}
		div.content[showing] {
			display: block;
		}
		.details-content {
			position: absolute;
			z-index: 10;
			right: 4px;
			padding: 4px;
			border: 1px solid black;
			border-radius: 4px 0 4px 4px;
			background: white;
			font-weight: initial;
		}
		div#index {
			display: none;
		}
		div#index.showing {
			display: block;
		}
	</style>
</head>
<body>
`)
	hb.WriteString(`<div class="help"><div class="title"><span class="title">Index</span><div id="index" style="float:right;font-size:initial;">`)
	hb.WriteString(`<details><summary>Index</summary><div class="details-content">`)
	contents[Index].asHTML(&hb, true)
	hb.WriteString(`</div></details>`)
	hb.WriteString(`</div></div><div id="0" class="content" showing>`)
	contents[Index].asHTML(&hb, true)
	hb.WriteString(`</div>`)
	for t, c := range contents {
		if t != Index {
			hb.WriteString(`<div class="content" id="` + strconv.Itoa(int(t)) + `">`)
			c.asHTML(&hb, true)
			hb.WriteString(`</div>`)
		}
	}
	hb.WriteString(`</div>`)
	hb.WriteString(`
	<script>
		const index = {`)
	for t := range contents {
		hb.WriteString(fmt.Sprintf(`%q:%q,`, strconv.Itoa(int(t)), t.String()))
	}
	hb.WriteString(`};
		function showTopic(id) {
			const show = index[id];
			if (show) {
				document.title = "GoGoL - " + show;
				document.querySelector("span.title").innerHTML = show;
				document.querySelectorAll("div.content").forEach(element => {
					element.toggleAttribute("showing", element.id == id);
				});
				const index = document.getElementById("index");
				if (id == 0) {
					index.classList.remove("showing");
				} else {
					index.classList.add("showing");
				}
			}
		}
		function hashChanged() {
			var id = location.hash.slice(1);
			if (!id) {
				id = 0;
			}
			showTopic(id);
		}
		window.addEventListener("hashchange", hashChanged);
		const id = location.hash.slice(1);
		if (id) {
			showTopic(id);
		}
	</script>
`)
	hb.WriteString(`</body></html>`)
	gtx.Execute(clipboard.WriteCmd{
		Type: "text/html",
		Data: io.NopCloser(strings.NewReader(hb.String())),
	})
}

func copyHtml(gtx layout.Context, t Topic, c content) {
	var hb strings.Builder
	hb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>GoGoL `)
	hb.WriteString(t.String())
	hb.WriteString(`</title>` + htmlStyling + `
</head>
<body>
`)
	c.asHTML(&hb, false)
	hb.WriteString(`</body></html>`)
	gtx.Execute(clipboard.WriteCmd{
		Type: "text/html",
		Data: io.NopCloser(strings.NewReader(hb.String())),
	})
}

func safeHtml(s string) string {
	return strings.ReplaceAll(html.EscapeString(s), "\n", "<br>")
}

type htmlBuilder interface {
	asHTML(hb *strings.Builder, full bool)
}

func asHTML(item any, hb *strings.Builder, full bool) {
	switch it := item.(type) {
	case htmlBuilder:
		it.asHTML(hb, full)
	case string:
		hb.WriteString(safeHtml(it))
	case []any:
		content(it).asHTML(hb, full)
	case fmt.Stringer:
		hb.WriteString(safeHtml(it.String()))
	}
}

func (c content) asHTML(hb *strings.Builder, full bool) {
	for _, item := range c {
		asHTML(item, hb, full)
	}
}

func (t Topic) asHTML(hb *strings.Builder, full bool) {
	txt := strings.TrimSuffix(t.String(), " Help")
	if full {
		hb.WriteString(`<a class="link" href="#` + strconv.Itoa(int(t)) + `">`)
		hb.WriteString(safeHtml(txt))
		hb.WriteString(`</a>`)
		return
	}
	hb.WriteString(`<span class="link">`)
	hb.WriteString(safeHtml(txt))
	hb.WriteString(`</span>`)
}

func (t topicLink) asHTML(hb *strings.Builder, full bool) {
	txt := t.Text
	if txt == "" {
		txt = strings.TrimSuffix(t.Topic.String(), " Help")
	}
	if full {
		hb.WriteString(`<a class="link" href="#` + strconv.Itoa(int(t.Topic)) + `">`)
		if t.Italic {
			hb.WriteString(`<em>`)
			hb.WriteString(safeHtml(txt))
			hb.WriteString(`</em>`)
		} else {
			hb.WriteString(safeHtml(txt))
		}
		hb.WriteString(`</a>`)
		return
	}
	hb.WriteString(`<span class="link">`)
	if t.Italic {
		hb.WriteString(`<em>`)
		hb.WriteString(safeHtml(txt))
		hb.WriteString(`</em>`)
	} else {
		hb.WriteString(safeHtml(txt))
	}
	hb.WriteString(`</span>`)
}

func (e externalLink) asHTML(hb *strings.Builder, _ bool) {
	hb.WriteString(`<a href="`)
	hb.WriteString(safeHtml(e.Url))
	hb.WriteString(`">`)
	if len(e.Text) > 0 {
		hb.WriteString(safeHtml(e.Text))
	} else {
		hb.WriteString(safeHtml(e.Url))
	}
	hb.WriteString(`</a>`)
}

func (b bold) asHTML(hb *strings.Builder, _ bool) {
	hb.WriteString(`<strong>`)
	hb.WriteString(safeHtml(string(b)))
	hb.WriteString(`</strong>`)
}

func (i italic) asHTML(hb *strings.Builder, _ bool) {
	hb.WriteString(`<em>`)
	hb.WriteString(safeHtml(string(i)))
	hb.WriteString(`</em>`)
}

func (bi boldItalic) asHTML(hb *strings.Builder, _ bool) {
	hb.WriteString(`<strong><em>`)
	hb.WriteString(safeHtml(string(bi)))
	hb.WriteString(`</em></strong>`)
}

func (b button) asHTML(hb *strings.Builder, _ bool) {
	hb.WriteString(`<span class="button">`)
	hb.WriteString(safeHtml(string(b)))
	hb.WriteString(`</span>`)
}

func (k keys) asHTML(hb *strings.Builder, _ bool) {
	hb.WriteString(`<span class="key">`)
	hb.WriteString(safeHtml(k.String()))
	hb.WriteString(`</span>`)
}

func (c code) asHTML(hb *strings.Builder, _ bool) {
	hb.WriteString(`<code>`)
	hb.WriteString(safeHtml(string(c)))
	hb.WriteString(`</code>`)
}

func (c codeCopyable) asHTML(hb *strings.Builder, _ bool) {
	hb.WriteString(`<code>`)
	hb.WriteString(safeHtml(string(c)))
	hb.WriteString(`</code>`)
}

var htmlCodeEscaper = strings.NewReplacer(
	`&`, "&amp;",
	`<`, "&lt;",
	`>`, "&gt;",
	"\t", "  ",
)

func (cb codeBlock) asHTML(hb *strings.Builder, _ bool) {
	hb.WriteString(`<pre>`)
	hb.WriteString(htmlCodeEscaper.Replace(cb.code))
	hb.WriteString(`</pre>`)
}

func (hdr header) asHTML(hb *strings.Builder, full bool) {
	tag := "h" + strconv.Itoa(hdr.level)
	hb.WriteString(`<` + tag + spacingStyleAttribute(0, hdr.spaceBefore, hdr.spaceAfter) + `>`)
	if hdr.mono {
		code(hdr.text).asHTML(hb, full)
	} else {
		hb.WriteString(safeHtml(hdr.text))
	}
	hb.WriteString(`</` + tag + `>`)
}

func (e expandable) asHTML(hb *strings.Builder, full bool) {
	opened := ""
	if e.initialExpanded {
		opened = " open"
	}
	hb.WriteString(`<details` + spacingStyleAttribute(e.indent, e.spaceBefore, e.spaceAfter) + opened + `><summary>`)
	asHTML(e.title, hb, full)
	hb.WriteString(`</summary>`)
	asHTML(e.content, hb, full)
	hb.WriteString(`</details>`)
}

func (t table) asHTML(hb *strings.Builder, full bool) {
	cellPadding := ""
	if t.padding != 0 {
		cellPadding = "--cell-padding:" + strconv.Itoa(int(t.padding)) + "px"
	}
	if t.borders {
		hb.WriteString(`<table class="bordered"` + spacingStyleAttribute(t.indent, t.spaceBefore, t.spaceAfter, cellPadding) + `>`)
	} else {
		hb.WriteString(`<table` + spacingStyleAttribute(t.indent, t.spaceBefore, t.spaceAfter, cellPadding) + `>`)
	}
	hasHeader := slices.ContainsFunc(t.columns, func(col tableColumn) bool { return col.header != "" })
	if hasHeader {
		hb.WriteString(`<thead><tr>`)
		for _, col := range t.columns {
			st := make([]string, 0)
			if col.width != 0 {
				st = append(st, "width:"+strconv.Itoa(col.width)+"%")
			}
			switch col.align {
			case text.End:
				st = append(st, "text-align:right")
			case text.Middle:
				st = append(st, "text-align:center")
			}
			if len(st) > 0 {
				hb.WriteString(`<th style="` + strings.Join(st, ";") + `">`)
			} else {
				hb.WriteString(`<th>`)
			}
			hb.WriteString(safeHtml(col.header))
			hb.WriteString(`</th>`)
		}
		hb.WriteString(`</tr></thead>`)
	}
	hb.WriteString(`<tbody>`)
	for _, row := range t.rows {
		hb.WriteString(`<tr>`)
		for _, col := range row {
			hb.WriteString(`<td>`)
			asHTML(col, hb, full)
			hb.WriteString(`</td>`)
		}
		if padC := len(t.columns) - len(row); padC > 0 {
			for range padC {
				hb.WriteString(`<td></td>`)
			}
		}
		hb.WriteString(`</tr>`)
	}
	hb.WriteString(`</tbody></table>`)
}

func (i indent) asHTML(hb *strings.Builder, full bool) {
	hb.WriteString(`<div` + spacingStyleAttribute(i.indent, i.spaceBefore, i.spaceAfter) + `>`)
	asHTML(i.content, hb, full)
	hb.WriteString(`</div>`)
}

func (i iconText) asHTML(hb *strings.Builder, _ bool) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, i.image); err == nil {
		hb.WriteString(`<img src="data:image/png;base64,`)
		data := buf.Bytes()
		hb.WriteString(base64.StdEncoding.EncodeToString(data))
		hb.WriteString(`" style="width:1.2em;height:1.2em;vertical-align:text-bottom;">`)
	}
}

func (i Image) asHTML(hb *strings.Builder, _ bool) {
	if data, err := images.LoadRaw(i.name); err == nil {
		hb.WriteString(`<img style="display:block;" src="data:image/png;base64,`)
		hb.WriteString(base64.StdEncoding.EncodeToString(data))
		hb.WriteString(`"` + spacingStyleAttribute(i.indent, i.spaceBefore, i.spaceAfter))
		if i.width != 0 {
			hb.WriteString(` width="` + strconv.Itoa(i.width) + `"`)
		}
		if i.height != 0 {
			hb.WriteString(` height="` + strconv.Itoa(i.height) + `"`)
		}
		if i.width == 0 && i.height == 0 {
			if img, err := images.LoadImage(i.name); err == nil {
				b := img.Bounds()
				hb.WriteString(` width="` + strconv.Itoa(b.Dx()/2) + `" height="` + strconv.Itoa(b.Dy()/2) + `"`)
			}
		}
		hb.WriteString(`>`)
	}
}

func (a hanging) asHTML(hb *strings.Builder, full bool) {
	gap := ""
	if a.gap != 0 {
		gap = "gap:" + strconv.Itoa(a.gap) + "px"
	}
	hb.WriteString(`<div class="hanging"` + spacingStyleAttribute(a.indent, a.spaceBefore, a.spaceAfter, gap) + `>`)
	hb.WriteString(`<div>`)
	var sb strings.Builder
	asHTML(a.prefix, &sb, full)
	if before, found := strings.CutSuffix(sb.String(), " "); found {
		hb.WriteString(before)
		hb.WriteString(`&nbsp;`)
	} else {
		hb.WriteString(sb.String())
	}
	hb.WriteString(`</div><div>`)
	asHTML(a.content, hb, full)
	hb.WriteString(`</div></div>`)
}

func (s separator) asHTML(hb *strings.Builder, _ bool) {
	hb.WriteString("<hr>")
}

func spacingStyleAttribute(indent, spaceBefore, spaceAfter unit.Dp, added ...string) string {
	var sb strings.Builder
	for _, add := range added {
		if add != "" {
			sb.WriteString(add + ";")
		}
	}
	if indent > 0 {
		sb.WriteString("margin-left:" + strconv.Itoa(int(indent)) + "px;")
	}
	if spaceBefore > 0 {
		sb.WriteString("margin-top:" + strconv.Itoa(int(spaceBefore)) + "px;")
	}
	if spaceAfter > 0 {
		sb.WriteString("margin-bottom:" + strconv.Itoa(int(spaceAfter)) + "px;")
	}
	if sb.Len() > 0 {
		return ` style="` + sb.String() + `"`
	}
	return ""
}
