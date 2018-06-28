package porcelain

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// see
// https://git-scm.com/docs/git-status#_changed_tracked_entries

// Line                                     Notes
// ------------------------------------------------------------
// # branch.oid <commit> | (initial)        Current commit.
// # branch.head <branch> | (detached)      Current branch.
// # branch.upstream <upstream_branch>      If upstream is set.
// # branch.ab +<ahead> -<behind>           If upstream is set and
// 	     the commit is present.
// ------------------------------------------------------------
type PorcInfo struct {
	branch string
	commit string
	// remote   string
	upstream string
	ahead    int
	behind   int

	untracked int
	unmerged  int

	Unstaged GitArea
	Staged   GitArea

	Entries []*Entry

	Changes   []Entry
	Renames   []Entry
	Unmerged  []Entry
	Untracked []Entry
	Ignored   []Entry
}

type GitArea struct {
	modified int
	added    int
	deleted  int
	renamed  int
	copied   int
}

type EntryState int

const (
	Unmodified EntryState = iota
	Modified
	Added
	Deleted
	Renamed
	Copied
	Updated // but unmerged
	Untracked
	Ignored
)

// X          Y     Meaning
// -------------------------------------------------
//           [MD]   not updated
// M        [ MD]   updated in index
// A        [ MD]   added to index
// D         [ M]   deleted from index
// R        [ MD]   renamed in index
// C        [ MD]   copied in index
// [MARC]           index and work tree matches
// [ MARC]     M    work tree changed since index
// [ MARC]     D    deleted in work tree
// -------------------------------------------------
// D           D    unmerged, both deleted
// A           U    unmerged, added by us
// U           D    unmerged, deleted by them
// U           A    unmerged, added by them
// D           U    unmerged, deleted by us
// A           A    unmerged, both added
// U           U    unmerged, both modified
// -------------------------------------------------
// ?           ?    untracked
// !           !    ignored
// -------------------------------------------------
type Entry struct {
	ChangeType   string
	XY           string
	SubModule    string
	ModeHead     string // stage 1
	ModeIndex    string // stage 2
	ModeStage3   string // stage 3
	ModeWorkTree string
	HashHead     string // stage 1
	HashIndex    string // stage 2
	HashStage3   string
	CopyScore    string
	Path         string
	OrigPath     string
}

// Field       Meaning
// --------------------------------------------------------
// <XY>        A 2 character field containing the staged and
// unstaged XY values described in the short format,
// with unchanged indicated by a "." rather than
// a space.
// <sub>       A 4 character field describing the submodule state.
// "N..." when the entry is not a submodule.
// "S<c><m><u>" when the entry is a submodule.
// <c> is "C" if the commit changed; otherwise ".".
// <m> is "M" if it has tracked changes; otherwise ".".
// <u> is "U" if there are untracked changes; otherwise ".".
// <mH>        The octal file mode in HEAD.
// <mI>        The octal file mode in the index.
// <mW>        The octal file mode in the worktree.
// <hH>        The object name in HEAD.
// <hI>        The object name in the index.
// <X><score>  The rename or copy score (denoting the percentage
// of similarity between the source and target of the
// move or copy). For example "R100" or "C75".
// <path>      The pathname.  In a renamed/copied entry, this
// is the path in the index and in the working tree.
// <sep>       When the `-z` option is used, the 2 pathnames are separated
// with a NUL (ASCII 0x00) byte; otherwise, a tab (ASCII 0x09)
// byte separates them.
// <origPath>  The pathname in the commit at HEAD.  This is only
// present in a renamed/copied entry, and tells
// where the renamed/copied contents came from.
// --------------------------------------------------------

func consumeNext(s *bufio.Scanner) string {
	if s.Scan() {
		return s.Text()
	}
	return ""
}

func (pi *PorcInfo) ParsePorcInfo(r io.Reader) error {
	var err error
	var s = bufio.NewScanner(r)

	for s.Scan() {
		if len(s.Text()) < 1 {
			continue
		}

		pi.ParseLine(s.Text())
	}

	return err
}

func (pi *PorcInfo) ParseLine(line string) error {
	s := bufio.NewScanner(strings.NewReader(line))
	// switch to a word based scanner
	s.Split(bufio.ScanWords)

	for s.Scan() {
		var err error
		var en *Entry
		text := s.Text()
		if text != "#" {
			en = &Entry{}
			pi.Entries = append(pi.Entries, en)
		}
		switch text {
		case "#":
			err = pi.parseBranchInfo(s)
		case "1":
			en.ChangeType = "1"
			err = pi.parseTrackedFile(s, en)
		case "2":
			en.ChangeType = "2"
			err = pi.parseRenamedFile(s, en)
		case "u":
			en.ChangeType = "u"
			err = pi.parseUnmergedFile(s, en)
			pi.unmerged++
		case "?":
			en.ChangeType = "?"
			err = pi.parseUntracked(s, en)
			pi.untracked++
		case "!":
			en.ChangeType = "!"
			err = pi.parseUntracked(s, en)
			pi.untracked++
		default:
			return fmt.Errorf("unexpected token (line) %v", text)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (pi *PorcInfo) parseBranchInfo(s *bufio.Scanner) (err error) {
	// uses the word based scanner from ParseLine
	for s.Scan() {
		switch s.Text() {
		case "branch.oid":
			pi.commit = consumeNext(s)
		case "branch.head":
			pi.branch = consumeNext(s)
		case "branch.upstream":
			pi.upstream = consumeNext(s)
		case "branch.ab":
			err = pi.parseAheadBehind(s)
		}
	}
	return err
}

func (pi *PorcInfo) parseAheadBehind(s *bufio.Scanner) error {
	// uses the word based scanner from ParseLine
	for s.Scan() {
		i, err := strconv.Atoi(s.Text()[1:])
		if err != nil {
			return err
		}

		switch s.Text()[:1] {
		case "+":
			pi.ahead = i
		case "-":
			pi.behind = i
		}
	}
	return nil
}

// parseTrackedFile parses the porcelain v2 output for tracked entries
// doc: https://git-scm.com/docs/git-status#_changed_tracked_entries
//
func (pi *PorcInfo) parseTrackedFile(s *bufio.Scanner, en *Entry) error {
	// uses the word based scanner from ParseLine
	var index int
	for s.Scan() {
		switch {
		case index == 0: // xy
			en.XY = s.Text()
			pi.parseXY(en.XY, en)
		case index == 1: // sub
			en.SubModule = s.Text()
		case index == 2: // mH - octal file mode in HEAD
			en.ModeHead = s.Text()
		case index == 3: // mI - octal file mode in index
			en.ModeIndex = s.Text()
		case index == 4: // mW - octal file mode in worktree
			en.ModeWorkTree = s.Text()
		case index == 5: // hH - object name in HEAD
			en.HashHead = s.Text()
		case index == 6: // hI - object name in index
			en.HashIndex = s.Text()
		case en.ChangeType == "1" && index == 7: // path
			en.Path = s.Text()
		case en.ChangeType == "2" && index == 7: // path
			en.CopyScore = s.Text()
		case en.ChangeType == "2" && index == 8: // path
			en.Path = s.Text()
		case en.ChangeType == "2" && index == 9:
			en.OrigPath = s.Text()
		default:
			return fmt.Errorf("unexpected token %v", s.Text())
		}
		index++
	}
	if en.ChangeType == "1" && index != 7 {
		return fmt.Errorf("left over tokens (change)")
	}
	if en.ChangeType == "2" && index != 9 {
		return fmt.Errorf("left over tokens (rename)")
	}

	return nil
}

func (pi *PorcInfo) parseUnmergedFile(s *bufio.Scanner, en *Entry) error {
	// uses the word based scanner from ParseLine
	var index int
	for s.Scan() {
		switch {
		case index == 0: // xy
			en.XY = s.Text()
			pi.parseXY(en.XY, en)
		case index == 1: // sub
			en.SubModule = s.Text()
		case index == 2: // mH - octal file mode in HEAD
			en.ModeHead = s.Text()
		case index == 3: // mI - octal file mode in index
			en.ModeIndex = s.Text()
		case index == 4: // mW - octal file mode in worktree
			en.ModeStage3 = s.Text()
		case index == 5: // mW - octal file mode in worktree
			en.ModeWorkTree = s.Text()
		case index == 6: // hH - object name in HEAD
			en.HashHead = s.Text()
		case index == 7: // hI - object name in index
			en.HashIndex = s.Text()
		case index == 8: // hI - object name in index
			en.HashStage3 = s.Text()
		case index == 9: // path
			en.Path = s.Text()
		default:
			return fmt.Errorf("unexpected token (unmerged) %v", s.Text())
		}
		index++
	}
	if index != 9 {
		return fmt.Errorf("left over tokens (unmerged)")
	}
	return nil
}

func (pi *PorcInfo) parseUntracked(s *bufio.Scanner, en *Entry) error {
	// uses the word based scanner from ParseLine
	var index int
	for s.Scan() {
		switch {
		case index == 0: // path
			en.Path = s.Text()
		default:
			return fmt.Errorf("unexpected token (untracked) %v", s.Text())
		}
		index++
	}
	return nil
}

func (pi *PorcInfo) parseXY(xy string, en *Entry) error {
	switch xy[:1] { // parse staged
	case "M":
		pi.Staged.modified++
	case "A":
		pi.Staged.added++
	case "D":
		pi.Staged.deleted++
	case "R":
		pi.Staged.renamed++
	case "C":
		pi.Staged.copied++
	}

	switch xy[1:] { // parse unstaged
	case "M":
		pi.Unstaged.modified++
	case "A":
		pi.Unstaged.added++
	case "D":
		pi.Unstaged.deleted++
	case "R":
		pi.Unstaged.renamed++
	case "C":
		pi.Unstaged.copied++
	}
	return nil
}

func (pi *PorcInfo) parseRenamedFile(s *bufio.Scanner, en *Entry) error {
	return pi.parseTrackedFile(s, en)
}
