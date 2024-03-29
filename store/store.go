package store

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/dav-m85/xbellum/xbel"
)

type version struct {
	id      string
	created time.Time
	xb      *xbel.XBEL
}

type Store struct {
	increment int
	// mu        sync.Mutex
	versions []version
	root     string
}

func NewStore(root string) *Store {
	reg := regexp.MustCompile(`^bkm_(\d{6}).xbel$`)
	fs, err := ioutil.ReadDir(root)
	if err != nil {
		panic(err)
	}
	st := Store{
		increment: -1,
		versions:  make([]version, len(fs)),
		root:      root,
	}
	for _, f := range fs {
		if m := reg.FindStringSubmatch(f.Name()); m != nil {
			re, err := ioutil.ReadFile(filepath.Join(root, f.Name()))
			if err != nil {
				panic(err)
			}
			xb, err := xbel.Parse(re)
			if err != nil {
				panic(err)
			}
			inc, err := strconv.Atoi(m[1])
			if err != nil {
				panic(err)
			}
			st.versions[inc] = version{
				id:      f.Name(),
				xb:      xb,
				created: f.ModTime(),
			}
			if inc > st.increment {
				st.increment = inc
			}
		}
	}
	log.Printf("Store increment:%d", st.increment)
	return &st
}

func (s *Store) get() *xbel.XBEL {
	if len(s.versions) == 0 {
		return nil
	}
	return s.versions[len(s.versions)-1].xb
}

func (s *Store) Get() ([]byte, error) {
	x := s.get()
	if x == nil {
		return nil, fmt.Errorf("no version available")
	}
	b := bytes.NewBuffer([]byte{})
	xbel.Write(b, x)
	return b.Bytes(), nil
}

func (s *Store) Set(d []byte) error {
	s.increment++
	fn := filepath.Join(s.root, fmt.Sprintf("bkm_%06d.xbel", s.increment))
	log.Printf("Store increment:%d", s.increment)

	s.versions = append(s.versions, version{
		id:      fn,
		xb:      xbel.MustParse(d),
		created: time.Now(),
	})
	// TODO If different from actual, record new instance
	return ioutil.WriteFile(fn, d, 0666)
}

func (s *Store) DiffAll() ([]Diff, error) {
	var diffs []Diff
	if len(s.versions) == 0 {
		return diffs, nil
	}
	parent := s.versions[0]
	for _, v := range s.versions[1:] {
		// Compare parent and v
		vb := xbel.Bookmarks(v.xb)
		pb := xbel.Bookmarks(parent.xb)
		added, removed := xbel.Diff(vb, pb)
		if len(added) > 0 || len(removed) > 0 {
			diffs = append(diffs, Diff{
				Version:       v.id,
				ParentVersion: parent.id,
				At:            v.created,
				Adds:          added,
				Removes:       removed,
			})
		}
		parent = v
	}
	return diffs, nil
}

type Diff struct {
	Version       string
	ParentVersion string
	At            time.Time
	Adds          []*xbel.Bookmark
	Removes       []*xbel.Bookmark
}
