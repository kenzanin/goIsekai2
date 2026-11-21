package tools

import "testing"

const htmlFixture = `<!doctype html><html><body>
<div id="main" class="card featured"><h1>Hello</h1><a href="/one" class="link">One</a></div>
<div class="card"><a href="/two" class="link">Two</a></div>
<img src="/cover.png" alt="cover">
</body></html>`

func TestHTMLSelect(t *testing.T) {
	got, err := Select(htmlFixture, ".link")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "One" || got[1] != "Two" {
		t.Fatalf("select = %v", got)
	}
	got, err = Select(htmlFixture, "#main")
	if err != nil || len(got) != 1 || got[0] != "HelloOne" {
		t.Fatalf("id select = %v, %v", got, err)
	}
}

func TestHTMLAttrAndXPath(t *testing.T) {
	got, err := Attr(htmlFixture, ".link", "href")
	if err != nil || got != "/one" {
		t.Fatalf("attr = %q, %v", got, err)
	}
	gotSlice, err := XPath(htmlFixture, `//a[@class='link']`)
	if err != nil || len(gotSlice) != 2 {
		t.Fatalf("xpath = %v, %v", gotSlice, err)
	}
	if _, err := Attr("<p>x</p>", ".missing", "href"); err == nil {
		t.Fatal("missing element should error")
	}
}
