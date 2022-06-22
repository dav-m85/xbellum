package xbel

import (
	"encoding/xml"
	"errors"
	"io"
	"sort"
)

const SUPPORTED_VERSION = "1.0"

const header = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE xbel PUBLIC "+//IDN python.org//DTD XML Bookmark Exchange Language 1.0//EN//XML" "http://pyxml.sourceforge.net/topics/dtds/xbel.dtd">
`

type XBEL struct {
	XMLName xml.Name `xml:"xbel"`
	Version string   `xml:"version,attr"`
	Folders []Folder `xml:"folder"`
}

type Folder struct {
	Title     string     `xml:"title"`
	ID        int        `xml:"id,attr"`
	Folders   []Folder   `xml:"folder"`
	Bookmarks []Bookmark `xml:"bookmark"`
}

type Bookmark struct {
	Title string `xml:"title"`
	ID    int    `xml:"id,attr"`
	Href  string `xml:"href,attr"`
}

func MustParse(buf []byte) *XBEL {
	xbel, err := Parse(buf)
	if err != nil {
		panic(err)
	}
	return xbel
}

func Parse(buf []byte) (*XBEL, error) {
	xbel := new(XBEL)
	err := xml.Unmarshal(buf, &xbel)
	if err == nil && xbel.Version != SUPPORTED_VERSION {
		return nil, errors.New("unsupported XBEL version")
	}
	return xbel, err
}

func Write(out io.Writer, nx *XBEL) {
	output, err := xml.MarshalIndent(nx, "  ", "    ")
	if err != nil {
		panic(err)
	}

	out.Write([]byte(header))
	out.Write(output)
}

// output, err := xml.MarshalIndent(nx, "  ", "    ")

// Filter keeps the passed Bookmark when true
type Filter func(b *Bookmark) bool

func Bookmarks(x *XBEL) (res []*Bookmark) {
	Walk(x, func(b *Bookmark) bool {
		nb := *b
		res = append(res, &nb)
		return true // false yeilds weird results on limits
	})
	return
}

// Diff does a dual exclusion, returning only bookmarks that are unique to
// a, or b.
func Diff(a, b []*Bookmark) (onlyA, onlyB []*Bookmark) {
	am := make(map[string]*Bookmark)
	bm := make(map[string]*Bookmark)
	for _, x := range a {
		nx := *x
		am[x.Href] = &nx // am contains all a
	}
	for _, x := range b {
		nx := *x
		bm[x.Href] = &nx // bm contains all b
	}
	for k, x := range bm {
		if _, ok := am[k]; !ok {
			onlyB = append(onlyB, x)
		}
	}
	for k, x := range am {
		if _, ok := bm[k]; !ok {
			onlyA = append(onlyA, x)
		}
	}
	sort.Sort(sortByHref(onlyA))
	sort.Sort(sortByHref(onlyB))
	return
}

type sortByHref []*Bookmark

func (a sortByHref) Len() int           { return len(a) }
func (a sortByHref) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortByHref) Less(i, j int) bool { return a[i].Href < a[j].Href }

// Walk modifies passed XBEL file with a Filter
func Walk(x *XBEL, filter Filter) *XBEL {
	var nf []Folder
	for _, folder := range x.Folders {
		y := walkFolder(folder, filter)
		// Skip empty folders
		if len(y.Bookmarks) == 0 && len(y.Folders) == 0 {
			continue
		}
		nf = append(nf, y)
	}
	return &XBEL{
		XMLName: x.XMLName,
		Version: SUPPORTED_VERSION,
		Folders: nf,
	}
}

// walkFolder apply filter recursively to all bookmarks within
func walkFolder(folder Folder, filter Filter) Folder {
	var nb []Bookmark
	for _, b := range folder.Bookmarks {
		if filter(&b) {
			nb = append(nb, b)
		}
	}
	var nf []Folder
	for _, child := range folder.Folders {
		y := walkFolder(child, filter)
		// Skip empty folders
		if len(y.Bookmarks) == 0 && len(y.Folders) == 0 {
			continue
		}
		nf = append(nf, y)
	}
	return Folder{
		Folders:   nf,
		Bookmarks: nb,
		Title:     folder.Title,
		ID:        folder.ID,
	}
}
