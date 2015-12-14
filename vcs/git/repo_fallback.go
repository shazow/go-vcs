package git

import (
	"github.com/shazow/go-vcs/vcs"
	"github.com/shazow/go-vcs/vcs/gitcmd"
)

// Repository features which depend on gitcmd fallback live in this file.

// TODO: Remove gitcmd fallback: Clone
func Clone(url, dir string, opt vcs.CloneOpt) (*Repository, error) {
	// FIXME: This will call Open ~3 times as it jumps between
	// gitcmd -> gogit -> gitcmd until we replace it with a native version or
	// refactor.
	_, err := gitcmd.Clone(url, dir, opt)
	if err != nil {
		return nil, err
	}
	return Open(dir)
}

// TODO: Remove gitcmd fallback: Repository.BlameFile
func (r *Repository) BlameFile(path string, opt *vcs.BlameOptions) ([]*vcs.Hunk, error) {
	return r.fallback.BlameFile(path, opt)
}

// TODO: Remove gitcmd fallback: Repository.Committers
func (r *Repository) Committers(opt vcs.CommittersOptions) ([]*vcs.Committer, error) {
	return r.fallback.Committers(opt)
}

// TODO: Remove gitcmd fallback: Repository.Diff
func (r *Repository) Diff(base, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	return r.fallback.Diff(base, head, opt)
}

// TODO: Remove gitcmd fallback: Repository.CrossRepoDiff
func (r *Repository) CrossRepoDiff(base vcs.CommitID, headRepo vcs.Repository, head vcs.CommitID, opt *vcs.DiffOptions) (*vcs.Diff, error) {
	return r.fallback.CrossRepoDiff(base, headRepo, head, opt)
}

// TODO: Remove gitcmd fallback: Repository.CrossRepoMergeBase
func (r *Repository) CrossRepoMergeBase(a vcs.CommitID, repoB vcs.Repository, b vcs.CommitID) (vcs.CommitID, error) {
	return r.fallback.CrossRepoMergeBase(a, repoB, b)
}

// TODO: Remove gitcmd fallback: Repository.Search
func (r *Repository) Search(at vcs.CommitID, opt vcs.SearchOptions) ([]*vcs.SearchResult, error) {
	return r.fallback.Search(at, opt)
}

// TODO: Remove gitcmd fallback: Repository.MergeBase
func (r *Repository) MergeBase(a, b vcs.CommitID) (vcs.CommitID, error) {
	return r.fallback.MergeBase(a, b)
}

// TODO: Remove gitcmd fallback: Repository.UpdateEverything
func (r *Repository) UpdateEverything(opt vcs.RemoteOpts) (*vcs.UpdateResult, error) {
	return r.fallback.UpdateEverything(opt)
}

// ResolveRevision returns the revision that the given revision
// specifier resolves to, or a non-nil error if there is no such
// revision.
// TODO: Remove partial gitcmd fallback: Repository.ResolveRevision
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
	// Do an extra lookup just in case it's a complex syntax we don't support
	ci, err = r.fallback.ResolveRevision(spec)
	if err == nil {
		return ci, nil
	}
	return ci, vcs.ErrRevisionNotFound
}
