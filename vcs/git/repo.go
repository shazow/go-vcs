package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shazow/go-git"
	"golang.org/x/tools/godoc/vfs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/gitcmd"
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
//
// This implementation does not provide any locking. Concurrency concerns
// should be handled by the consumer of this library.
type Repository struct {
	repo *git.Repository
	// fallback is used for features that are not implemented here yet.
	fallback *gitcmd.Repository
}

func (r *Repository) RepoDir() string {
	return r.repo.Path
}

func (r *Repository) GitRootDir() string {
	return r.RepoDir()
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
		return nil, &os.PathError{
			Op:   fmt.Sprintf("Open git repo [%s]", err.Error()),
			Path: dir,
			Err:  os.ErrNotExist,
		}
	}

	return &Repository{
		repo:     repo,
		fallback: &gitcmd.Repository{Dir: dir},
	}, nil
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

	var behindAhead *git.Commit
	var mergedInto *git.Commit
	var branches []*vcs.Branch

	if opt.BehindAheadBranch != "" {
		behindAhead, err = r.repo.GetCommitOfBranch(opt.BehindAheadBranch)
		if err != nil {
			return nil, err
		}
	}

	if opt.MergedInto != "" {
		mergedInto, err = r.repo.GetCommit(opt.MergedInto)
		if err != nil {
			return nil, err
		}
	}

	for _, name := range names {
		id, err := r.ResolveBranch(name)
		if err != nil {
			return nil, err
		}
		branch := &vcs.Branch{Name: name, Head: id}
		if !opt.IncludeCommit && opt.ContainsCommit == "" && opt.MergedInto == "" && opt.BehindAheadBranch == "" {
			// Short circuit fetching the commit and use a minimal branch object.
			branches = append(branches, branch)
			continue
		}

		commit, err := r.repo.GetCommit(string(id))
		if err != nil {
			return nil, err
		}
		if opt.IncludeCommit {
			branch.Commit = r.vcsCommit(commit)
		}

		commitId := commit.Id.String()
		if opt.ContainsCommit != "" && opt.ContainsCommit != commitId {
			if !commit.IsAncestor(opt.ContainsCommit) {
				continue
			}
		}
		if opt.MergedInto != "" && opt.MergedInto != commitId {
			// MergedInto returns branches which fully contain the MergedInto
			// commit, which is the reverse traversal of ContainsCommit.
			if !mergedInto.IsAncestor(commitId) {
				continue
			}
		}
		if opt.BehindAheadBranch != "" {
			behind, ahead, _ := commit.BehindAhead(behindAhead.Id.String())
			branch.Counts = &vcs.BehindAhead{
				Behind: uint32(behind),
				Ahead:  uint32(ahead),
			}
		}
		branches = append(branches, branch)
	}

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

// Convert a git.Commit to a vcs.Commit
func (r *Repository) vcsCommit(commit *git.Commit) *vcs.Commit {
	var committer *vcs.Signature
	if commit.Committer != nil {
		committer = &vcs.Signature{
			Name:  commit.Committer.Name,
			Email: commit.Committer.Email,
			Date:  pbtypes.NewTimestamp(commit.Committer.When),
		}
	}

	n := commit.ParentCount()
	parentIds := commit.ParentIds()
	parents := make([]vcs.CommitID, 0, len(parentIds))
	for _, id := range parentIds {
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
	}
}

// GetCommit returns the commit with the given commit ID, or
// ErrCommitNotFound if no such commit exists.
func (r *Repository) GetCommit(commitID vcs.CommitID) (*vcs.Commit, error) {
	commit, err := r.repo.GetCommit(string(commitID))
	if err != nil {
		return nil, standardizeError(err)
	}

	return r.vcsCommit(commit), nil
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
