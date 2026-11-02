package chat

import "testing"

func TestRenderVisuals(t *testing.T) {
	cases := []struct{ in, want string }{
		{"<preview>barplot.png</preview>", "![](barplot.png)"},
		{"Here:\n<preview> barplot.png </preview>", "Here:\n![](barplot.png)"},
		{"<preview path=\"a.png\"></preview>", "![](a.png)"},
		{"plain text", "plain text"},
		{"```xml\n<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"10\"></svg>\n```",
			"```svg\n<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"10\"></svg>\n```"},
	}
	for _, c := range cases {
		if got := renderVisuals(c.in); got != c.want {
			t.Errorf("renderVisuals(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}