package reader

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

type containerDoc struct {
	Rootfiles []struct {
		FullPath string `xml:"full-path,attr"`
	} `xml:"rootfiles>rootfile"`
}

type opfDoc struct {
	Manifest []opfManifestItem `xml:"manifest>item"`
	Spine    opfSpine          `xml:"spine"`
}

type opfManifestItem struct {
	ID         string `xml:"id,attr"`
	Href       string `xml:"href,attr"`
	MediaType  string `xml:"media-type,attr"`
	Properties string `xml:"properties,attr"`
}

type opfSpine struct {
	TOC      string            `xml:"toc,attr"`
	Itemrefs []opfSpineItemref `xml:"itemref"`
}

type opfSpineItemref struct {
	IDRef string `xml:"idref,attr"`
}

type ncxDoc struct {
	NavMap ncxNavMap `xml:"navMap"`
}

type ncxNavMap struct {
	NavPoints []ncxNavPoint `xml:"navPoint"`
}

type ncxNavPoint struct {
	Label struct {
		Text string `xml:"text"`
	} `xml:"navLabel"`
	Content struct {
		Src string `xml:"src,attr"`
	} `xml:"content"`
	Children []ncxNavPoint `xml:"navPoint"`
}

func loadEPUB(pathName string) (*Document, error) {
	zr, err := zip.OpenReader(pathName)
	if err != nil {
		return nil, err
	}
	defer zr.Close()

	files := map[string][]byte{}
	for _, f := range zr.File {
		bytes, err := readZipFile(f)
		if err != nil {
			return nil, err
		}
		files[path.Clean(f.Name)] = bytes
	}

	containerBytes, ok := files[path.Clean("META-INF/container.xml")]
	if !ok {
		return nil, fmt.Errorf("invalid epub: missing META-INF/container.xml")
	}

	var container containerDoc
	if err := xml.Unmarshal(containerBytes, &container); err != nil {
		return nil, fmt.Errorf("invalid epub container: %w", err)
	}
	if len(container.Rootfiles) == 0 || strings.TrimSpace(container.Rootfiles[0].FullPath) == "" {
		return nil, fmt.Errorf("invalid epub: missing rootfile path")
	}

	opfPath := path.Clean(container.Rootfiles[0].FullPath)
	opfBytes, ok := files[opfPath]
	if !ok {
		return nil, fmt.Errorf("invalid epub: missing OPF %q", opfPath)
	}

	var opf opfDoc
	if err := xml.Unmarshal(opfBytes, &opf); err != nil {
		return nil, fmt.Errorf("invalid epub OPF: %w", err)
	}
	if len(opf.Manifest) == 0 || len(opf.Spine.Itemrefs) == 0 {
		return nil, fmt.Errorf("invalid epub: empty manifest or spine")
	}

	manifestByID := make(map[string]opfManifestItem, len(opf.Manifest))
	for _, item := range opf.Manifest {
		manifestByID[item.ID] = item
	}

	tocTitles := make(map[string]string)
	opfDir := path.Dir(opfPath)
	if opf.Spine.TOC != "" {
		if tocItem, found := manifestByID[opf.Spine.TOC]; found {
			tocPath := path.Clean(path.Join(opfDir, tocItem.Href))
			if tocBytes, ok := files[tocPath]; ok {
				for href, title := range parseNCXTitles(tocBytes, path.Dir(tocPath)) {
					tocTitles[href] = title
				}
			}
		}
	}

	for _, item := range opf.Manifest {
		if strings.Contains(item.Properties, "nav") {
			navPath := path.Clean(path.Join(opfDir, item.Href))
			if navBytes, ok := files[navPath]; ok {
				for href, title := range parseHTMLNavTitles(navBytes, path.Dir(navPath)) {
					tocTitles[href] = title
				}
			}
		}
	}

	doc := &Document{
		Title:     filepath.Base(pathName),
		Format:    "epub",
		Lines:     make([]string, 0, 8192),
		LineIndex: make([]LineAnchor, 0, 8192),
	}

	for spineIdx, ref := range opf.Spine.Itemrefs {
		item, ok := manifestByID[ref.IDRef]
		if !ok {
			continue
		}
		if !strings.Contains(item.MediaType, "html") && !strings.Contains(item.MediaType, "xhtml") {
			continue
		}

		spinePath := path.Clean(path.Join(opfDir, item.Href))
		body, ok := files[spinePath]
		if !ok {
			continue
		}

		extracted, err := extractHTMLText(body)
		if err != nil {
			return nil, fmt.Errorf("epub parse %q: %w", spinePath, err)
		}

		heading := strings.TrimSpace(tocTitles[normalizeHref(spinePath)])
		if heading == "" {
			heading = strings.TrimSpace(extracted.Heading)
		}
		if heading == "" {
			heading = fallbackChapterTitle(spinePath, spineIdx)
		}
		appendLine(doc, fmt.Sprintf("=== Chapter: %s ===", heading), LineAnchor{Page: -1, Spine: spineIdx, Offset: 0})

		offset := 1
		for _, line := range extracted.Lines {
			appendLine(doc, line, LineAnchor{Page: -1, Spine: spineIdx, Offset: offset})
			offset++
		}
	}

	if len(doc.Lines) == 0 {
		return nil, fmt.Errorf("invalid epub: no readable chapters")
	}

	return doc, nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}

func fallbackChapterTitle(spinePath string, spineIdx int) string {
	base := strings.TrimSpace(strings.TrimSuffix(path.Base(spinePath), path.Ext(spinePath)))
	if base != "" {
		return strings.ReplaceAll(base, "_", " ")
	}
	return fmt.Sprintf("Section %d", spineIdx+1)
}

func parseNCXTitles(data []byte, baseDir string) map[string]string {
	out := map[string]string{}
	var ncx ncxDoc
	if err := xml.Unmarshal(data, &ncx); err != nil {
		return out
	}

	var walk func(points []ncxNavPoint)
	walk = func(points []ncxNavPoint) {
		for _, p := range points {
			src := strings.TrimSpace(p.Content.Src)
			label := strings.TrimSpace(p.Label.Text)
			if src != "" && label != "" {
				full := normalizeHref(path.Clean(path.Join(baseDir, src)))
				if _, exists := out[full]; !exists {
					out[full] = label
				}
			}
			if len(p.Children) > 0 {
				walk(p.Children)
			}
		}
	}
	walk(ncx.NavMap.NavPoints)
	return out
}

func parseHTMLNavTitles(data []byte, baseDir string) map[string]string {
	out := map[string]string{}
	node, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return out
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			href := ""
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					href = strings.TrimSpace(attr.Val)
					break
				}
			}
			if href != "" {
				title := strings.TrimSpace(joinNodeText(n))
				if title != "" {
					full := normalizeHref(path.Clean(path.Join(baseDir, href)))
					if _, exists := out[full]; !exists {
						out[full] = title
					}
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return out
}

type htmlExtract struct {
	Lines   []string
	Heading string
}

func extractHTMLText(data []byte) (*htmlExtract, error) {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	body := findNode(doc, "body")
	if body == nil {
		body = doc
	}

	out := &htmlExtract{Lines: make([]string, 0, 256)}
	tokens := make([]string, 0, 32)

	flush := func() {
		if len(tokens) == 0 {
			return
		}
		line := strings.Join(tokens, " ")
		line = strings.Join(strings.Fields(strings.TrimSpace(line)), " ")
		if line != "" {
			out.Lines = append(out.Lines, line)
		}
		tokens = tokens[:0]
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.TextNode {
			text := strings.Join(strings.Fields(n.Data), " ")
			if text != "" {
				tokens = append(tokens, text)
			}
			return
		}
		if n.Type != html.ElementNode {
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				walk(child)
			}
			return
		}

		tag := strings.ToLower(n.Data)
		if isSkippableHTMLTag(tag) {
			return
		}

		if tag == "br" {
			flush()
			return
		}

		if isBlockHTMLTag(tag) {
			flush()
		}

		if out.Heading == "" && isHeadingHTMLTag(tag) {
			h := strings.Join(strings.Fields(joinNodeText(n)), " ")
			if h != "" {
				out.Heading = h
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}

		if isBlockHTMLTag(tag) {
			flush()
		}
	}

	walk(body)
	flush()
	if len(out.Lines) == 0 {
		out.Lines = []string{""}
	}
	return out, nil
}

func findNode(root *html.Node, tag string) *html.Node {
	if root == nil {
		return nil
	}
	if root.Type == html.ElementNode && root.Data == tag {
		return root
	}
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if found := findNode(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func joinNodeText(node *html.Node) string {
	parts := make([]string, 0, 8)
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				parts = append(parts, text)
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return strings.Join(parts, " ")
}

func normalizeHref(href string) string {
	trimmed := strings.TrimSpace(href)
	if idx := strings.Index(trimmed, "#"); idx >= 0 {
		trimmed = trimmed[:idx]
	}
	return path.Clean(trimmed)
}

func isSkippableHTMLTag(tag string) bool {
	switch tag {
	case "script", "style", "noscript", "svg", "math":
		return true
	default:
		return false
	}
}

func isHeadingHTMLTag(tag string) bool {
	switch tag {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		return true
	default:
		return false
	}
}

func isBlockHTMLTag(tag string) bool {
	switch tag {
	case "p", "div", "section", "article", "header", "footer", "main", "aside",
		"ul", "ol", "li", "blockquote", "pre", "table", "thead", "tbody", "tr", "td", "th",
		"h1", "h2", "h3", "h4", "h5", "h6", "nav":
		return true
	default:
		return false
	}
}
