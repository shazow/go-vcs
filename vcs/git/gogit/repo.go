package git

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shazow/go-git"
	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sqs/pbtypes"
)

func init() {
	// Overwrite the git opener to return repositories that use the
	// gogits native-go implementation.
	vcs.RegisterOpener("git", func(dir string) (vcs.Repository, error) {
		return Open(dir)
	})
}

// Repository is a git VCS repository.
type Repository struct {
	repo *git.Repository

	// TODO: Do we need locking?
}

func (r *Repository) RepoDir() string {
	return r.repo.Path
}

func (r *Repository) String() string {
	return fmt.Sprintf("git (gogit) repo at %s", r.RepoDir())
}

func Open(dir string) (*Repository, error) {
	if _, err := os.Stat(filepath.Join(dir, ".git")); !os.IsNotExist(err) {
		// Append .git to path
		dir = filepath.Join(dir, ".git")
	}

	repo, err := git.OpenRepository(dir)
	if err != nil {
		// FIXME: Wrap in vcs error?
		return nil, err
	}

	return &Repository{
		repo: repo,
	}, nil
}

// ResolveRevision returns the revision that the given revision
// specifier resolves to, or a non-nil error if there is no such
// revision.
func (r *Repository) ResolveRevision(spec string) (vcs.CommitID, error) {
	// TODO: git rev-parse supports a horde of complex syntaxes, it will be a fair bit more work to support all of them.
	// e.g. "master@{yesterday}", "master~3", and various text/path/tree traversal search.
	ci, err := r.ResolveTag(spec)
	if err == nil {
		return ci, nil
	}
	ci, err = r.ResolveBranch(spec)
	if err == nil {
		return ci, nil
	}
	return ci, vcs.ErrRevisionNotFound
}

// ResolveTag returns the tag with the given name, or
// ErrTagNotFound if no such tag exists.
func (r *Repository) ResolveTag(name string) (vcs.CommitID, error) {
	id, err := r.repo.GetCommitIdOfTag(name)
	if git.IsNotFound(err) {
		return "", vcs.ErrTagNotFound
	} else if err != nil {
		// Unexpected error
		return "", err
	}
	return vcs.CommitID(id), nil
}

// ResolveBranch returns the branch with the given name, or
// ErrBranchNotFound if no such branch exists.
func (r *Repository) ResolveBranch(name string) (vcs.CommitID, error) {
	id, err := r.repo.GetCommitIdOfBranch(name)
	if git.IsNotFound(err) {
		return "", vcs.ErrBranchNotFound
	} else if err != nil {
		// Unexpected error
		return "", err
	}
	return vcs.CommitID(id), nil
}

// Branches returns a list of all branches in the repository.
func (r *Repository) Branches(opt vcs.BranchesOptions) ([]*vcs.Branch, error) {
	names, err := r.repo.GetBranches()
	if err != nil {
		return nil, err
	}
	if opt.BehindAheadBranch != "" {
		return nil, fmt.Errorf("vcs.BranchesOptions BehindAheadBranch not implemented")
	}

	var branches []*vcs.Branch
	for _, name := range names {
		id, err := r.ResolveBranch(name)
		if err != nil {
			return nil, err
		}
		branch := &vcs.Branch{Name: name, Head: id}
		if opt.IncludeCommit {
			branch.Commit, err = r.GetCommit(id)
			if err != nil {
				return nil, err
			}
		}
		branches = append(branches, branch)
	}

	// TODO: opt.MergedInto
	// TODO: opt.ContainsCommit
	return branches, nil
}

// Tags returns a list of all tags in the repository.
func (r *Repository) Tags() ([]*vcs.Tag, error) {
	names, err := r.repo.GetTags()
	if err != nil {
		return nil, err
	}

	tags := make([]*vcs.Tag, 0, len(names))
	for _, name := range names {
		id, err := r.ResolveTag(name)
		if err != nil {
			return nil, err
		}
		tags = append(tags, &vcs.Tag{Name: name, CommitID: vcs.CommitID(id)})
	}

	return tags, nil
}

// GetCommit returns the commit with the given commit ID, or
// ErrCommitNotFound if no such commit exists.
func (r *Repository) GetCommit(commitID vcs.CommitID) (*vcs.Commit, error) {
	commit, err := r.repo.GetCommit(string(commitID))
	if err != nil {
		return nil, standardizeError(err)
	}

	var committer *vcs.Signature
	if commit.Committer != nil {
		committer = &vcs.Signature{
			Name:  commit.Committer.Name,
			Email: commit.Committer.Email,
			Date:  pbtypes.NewTimestamp(commit.Committer.When),
		}
	}

	n := commit.ParentCount()
	parents := make([]vcs.CommitID, 0, n)
	for i := 0; i < n; i++ {
		id, err := commit.ParentId(i)
		if err != nil {
			return nil, standardizeError(err)
		}
		parents = append(parents, vcs.CommitID(id.String()))
	}
	if n == 0 {
		// Required to make reflect.DeepEqual tests pass. :/
		parents = nil
	}

	return &vcs.Commit{
		ID: vcs.CommitID(commit.Id.String()),
		// TODO: Check nil on commit.Author?
		Author: vcs.Signature{
			Name:  commit.Author.Name,
			Email: commit.Author.Email,
			Date:  pbtypes.NewTimestamp(commit.Author.When),
		},
		Committer: committer,
		Message:   strings.TrimSuffix(commit.Message(), "\n"),
		Parents:   parents,
	}, nil
}

// Commits returns all commits matching the options, as well as
// the total number of commits (the count of which is not subject
// to the N/Skip options).
//
// Optionally, the caller can request the total not to be computed,
// as this can be expensive for large branches.
func (r *Repository) Commits(opt vcs.CommitsOptions) ([]*vcs.Commit, uint, error) {
	var total uint = 0
	var commits []*vcs.Commit
	var err error

	cur, err := r.repo.GetCommit(string(opt.Head))
	if err != nil {
		return commits, total, standardizeError(err)
	}

	parents := []*git.Commit{cur}
	for opt.N == 0 || opt.N > uint(len(commits)) {
		// Pop FIFO
		cur, parents = parents[len(parents)-1], parents[:len(parents)-1]
		if cur.Id.String() == string(opt.Base) {
			// FIXME: Is this the correct condition for opt.Base? Please review.
			break
		}

		if opt.Skip <= total {
			ci, err := r.GetCommit(vcs.CommitID(cur.Id.String()))
			if err != nil {
				return nil, 0, err
			}
			commits = append(commits, ci)
		}
		total++

		// Store all the parents
		for p, stop := 0, cur.ParentCount(); p < stop; p++ {
			pcommit, err := cur.Parent(p)
			if err != nil {
				return nil, 0, err
			}
			parents = append(parents, pcommit)
		}
		if len(parents) == 0 {
			break
		}
	}

	if opt.NoTotal {
		return commits, 0, err
	}

	for len(parents) > 0 {
		// Pop FIFO
		cur, parents = parents[len(parents)-1], parents[:len(parents)-1]
		total++

		// Store all the parents
		for p, stop := 0, cur.ParentCount(); p < stop; p++ {
			pcommit, err := cur.Parent(p)
			if err != nil {
				return nil, 0, err
			}
			parents = append(parents, pcommit)
		}
	}

	return commits, total, err
}

// Committers returns the per-author commit statistics of the repo.
func (r *Repository) Committers(committerOpts vcs.CommittersOptions) ([]*vcs.Committer, error) {
	return nil, errors.New("gogit: Committers not implemented")
}

// FileSystem opens the repository file tree at a given commit ID.
func (r *Repository) FileSystem(at vcs.CommitID) (vfs.FileSystem, error) {
	ci, err := r.repo.GetCommit(string(at))
	if err != nil {
		return nil, err
	}
	return &filesystem{
		dir:  r.repo.Path,
		oid:  string(at),
		tree: &ci.Tree,
		repo: r.repo,
	}, nil
}
