package xbel

import (
	"encoding/xml"
	"errors"
	"io"
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
