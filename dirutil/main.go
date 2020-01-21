package dirutil

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/cznic/mathutil"
	"github.com/runner-mei/ftp"
)

type DiffEntry struct {
	Name   string
	Size   uint64
	Time   time.Time
	Digset int
}

type DiffResult struct {
	Filename    string    `json:"filename"`
	Result      string    `json:"result"`
	LeftLastAt  time.Time `json:"left_last_at"`
	RightLastAt time.Time `json:"right_last_at"`
}

func Diff(left, right string) ([]DiffResult, error) {
	left_filenames, e := readDir(left)
	if nil != e {
		return nil, e
	}
	right_filenames, e := readDir(right)
	if nil != e {
		return nil, e
	}
	if 0 == len(left_filenames) && 0 == len(right_filenames) {
		return []DiffResult{}, nil
	}

	lefts := map[string]DiffEntry{}
	if len(left_filenames) > 0 {
		for _, en := range left_filenames {
			lefts[strings.ToLower(en.Name)] = en
		}
	}

	results := make([]DiffResult, 0, mathutil.Max(len(left_filenames), len(right_filenames)))
	if len(right_filenames) > 0 {
		for _, right_en := range right_filenames {
			left_en, ok := lefts[strings.ToLower(right_en.Name)]
			if ok {
				if left_en.Size == right_en.Size {
					ok, e := digsetEquals(left, left_en, right, right_en)
					if nil != e {
						results = append(results, DiffResult{
							Filename:    right_en.Name,
							RightLastAt: right_en.Time,
							LeftLastAt:  left_en.Time,
							Result:      "error:" + e.Error(),
						})
					} else if ok {
						results = append(results, DiffResult{
							Filename:    right_en.Name,
							RightLastAt: right_en.Time,
							LeftLastAt:  left_en.Time,
						})
					} else {
						results = append(results, DiffResult{
							Filename:    right_en.Name,
							Result:      "different",
							RightLastAt: right_en.Time,
							LeftLastAt:  left_en.Time,
						})
					}
				} else {
					results = append(results, DiffResult{
						Filename:    right_en.Name,
						Result:      "different",
						RightLastAt: right_en.Time,
						LeftLastAt:  left_en.Time,
					})
				}
				delete(lefts, strings.ToLower(right_en.Name))
			} else {
				results = append(results, DiffResult{
					Filename:    right_en.Name,
					Result:      "rightOnly",
					RightLastAt: right_en.Time,
				})
			}
		}
	}

	for _, left_en := range lefts {
		results = append(results, DiffResult{
			Filename:   left_en.Name,
			Result:     "leftOnly",
			LeftLastAt: left_en.Time,
		})
	}
	return results, nil
}

func digsetEquals(left string, left_en DiffEntry, right string, right_en DiffEntry) (bool, error) {
	if 0 == left_en.Digset || 0 == right_en.Digset {
		return true, nil
	}

	if 1 == left_en.Digset {
		if 1 == right_en.Digset || 3 == right_en.Digset {
			return fileEqual(left, left_en.Name, right, right_en.Name, ".md5")
		}
		return false, nil
	}

	if 2 == left_en.Digset {
		if 2 == right_en.Digset || 3 == right_en.Digset {
			return fileEqual(left, left_en.Name, right, right_en.Name, ".sha1")
		}
		return false, nil
	}

	if 1 == right_en.Digset || 3 == right_en.Digset {
		return fileEqual(left, left_en.Name, right, right_en.Name, ".md5")
	}
	return fileEqual(left, left_en.Name, right, right_en.Name, ".sha1")
}

func fileEqual(left, left_name, right, right_name, ext string) (bool, error) {
	left_txt, e := readFile(left, left_name+ext)
	if nil != e {
		return false, e
	}
	right_txt, e := readFile(right, right_name+ext)
	if nil != e {
		return false, e
	}
	return bytes.Equal(left_txt, right_txt), nil
}

func readDirFromOS(pa string) ([]DiffEntry, error) {
	f, e := os.Open(pa)
	if nil != e {
		return nil, e
	}
	defer f.Close()

	names, e := f.Readdirnames(-1)
	if nil != e {
		return nil, e
	}

	var results []DiffEntry
	for _, name := range names {
		st, e := os.Stat(name)
		if nil != e {
			return nil, e
		}

		results = append(results, DiffEntry{
			Name: name,
			Size: uint64(st.Size()),
			Time: st.ModTime()})
	}
	return results, nil
}

func readDir(url_str string) ([]DiffEntry, error) {
	if strings.HasPrefix(url_str, "//") {
		return readDirFromOS(url_str)
	}
	u, e := url.Parse(url_str)
	if nil != e {
		return nil, e
	}

	switch strings.ToLower(u.Scheme) {
	case "ftp":
	default:
		if "" == u.Scheme {
			return readDirFromOS(url_str)
		}
		return nil, errors.New("only supports ftp directory.")
	}

	conn, e := ftp.DialTimeout(u.Host, 10*time.Second)
	if nil != e {
		return nil, e
	}
	defer conn.Logout()

	if user := u.User.Username(); "" != user {
		password, _ := u.User.Password()
		if e = conn.Login(user, password); nil != e {
			return nil, e
		}
	}

	if "" != u.Path {
		if e = conn.ChangeDir(u.Path); nil != e {
			return nil, e
		}
	}
	entries, e := conn.List(u.Path)
	if nil != e {
		return nil, e
	}

	var results []DiffEntry
	if len(entries) > 0 {
		digsets := map[string]int{}
		results = make([]DiffEntry, 0, len(entries))
		for _, en := range entries {
			name := strings.ToLower(en.Name)
			if strings.HasSuffix(name, ".md5") {
				digsets[strings.TrimSuffix(name, ".md5")] = 1
				continue
			}
			if strings.HasSuffix(name, ".sha1") {
				if _, ok := digsets[strings.TrimSuffix(name, ".sha1")]; ok {
					digsets[strings.TrimSuffix(name, ".sha1")] = 3
				} else {
					digsets[strings.TrimSuffix(name, ".sha1")] = 2
				}
				continue
			}

			results = append(results, DiffEntry{
				Name: en.Name,
				Size: en.Size,
				Time: en.Time,
			})
		}

		for idx, _ := range results {
			if v, ok := digsets[strings.ToLower(results[idx].Name)]; ok {
				results[idx].Digset = v
			}
		}
	}
	return results, nil
}

func readFile(url_str, filename string) ([]byte, error) {
	u, e := url.Parse(url_str)
	if nil != e {
		return nil, e
	}

	switch strings.ToLower(u.Scheme) {
	case "ftp":
	default:
		return nil, errors.New("only supports ftp directory.")
	}

	conn, e := ftp.DialTimeout(u.Host, 10*time.Second)
	if nil != e {
		return nil, e
	}
	defer conn.Logout()

	if user := u.User.Username(); "" != user {
		password, _ := u.User.Password()
		if e = conn.Login(user, password); nil != e {
			return nil, e
		}
	}

	if "" != u.Path {
		if e = conn.ChangeDir(u.Path); nil != e {
			return nil, e
		}
	}

	reader, e := conn.Retr(filename)
	if nil != e {
		return nil, e
	}
	defer reader.Close()

	return ioutil.ReadAll(reader)
}
