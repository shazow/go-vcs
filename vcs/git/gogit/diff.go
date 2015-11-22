package git

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/shazow/go-git"
)

type diffEntry struct {
	id    string
	path  string
	entry *git.TreeEntry
}

type differ struct {
	*diffmatchpatch.DiffMatchPatch
	aPrefix string
	bPrefix string
}

func (diff *differ) WriteHeader(w io.Writer, a, b diffEntry) error {
	var aPath, bPath string
	if a.path == "" && b.path == "" {
		return errors.New("diff header: no path for both versions")
	}
	if a.path != "" {
		aPath = filepath.Join(diff.aPrefix, a.path)
		bPath = aPath
	}
	if b.path != "" {
		bPath = filepath.Join(diff.bPrefix, b.path)
		if aPath == "" {
			aPath = bPath
		}
	}
	// TODO: Detect renames?

	fmt.Fprintf(w, "diff --git %s %s\n", aPath, bPath)
	if a.entry == nil {
		fmt.Fprintf(w, "new file mode %o\n", b.entry.EntryMode())
		fmt.Fprintf(w, "index 0000000000000000000000000000000000000000..%s\n", b.id)
		fmt.Fprintf(w, "--- /dev/null\n")
		fmt.Fprintf(w, "+++ %s\n", bPath)
	} else if b.entry == nil {
		fmt.Fprintf(w, "deleted file mode %o\n", a.entry.EntryMode())
		fmt.Fprintf(w, "index %s..0000000000000000000000000000000000000000\n", a.id)
		fmt.Fprintf(w, "--- %s\n", aPath)
		fmt.Fprintf(w, "+++ /dev/null\n")
	} else {
		fmt.Fprintf(w, "index %s..%s %o\n", a.id, b.id, b.entry.EntryMode())
		fmt.Fprintf(w, "--- %s\n", aPath)
		fmt.Fprintf(w, "+++ %s\n", bPath)
	}
	return nil
}

func (diff *differ) Write(w io.Writer, a, b diffEntry) error {
	var aBody, bBody []byte
	var err error
	if a.entry != nil {
		aBody, err = readEntry(a.entry)
	}
	if err != nil {
		return err
	}
	if b.entry != nil {
		bBody, err = readEntry(b.entry)
	}
	if err != nil {
		return err
	}
	if err = diff.WriteHeader(w, a, b); err != nil {
		return err
	}

	patch := diff.PatchMake(string(aBody), string(bBody))
	diffText := strings.Replace(diff.PatchToText(patch), "%0A", "", -1)
	// PatchToText conveniently adds html-encoded line breaks :/
	_, err = fmt.Fprint(w, diffText)
	return err
}
