package vcs_test

import (
	"testing"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
)

func TestMerger_MergeBase(t *testing.T) {
	t.Parallel()

	// TODO(sqs): implement for hg
	// TODO(sqs): make a more complex test case

	cmds := []string{
		"echo line1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testbase",
		"git checkout -b b2",
		"echo line2 >> f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git checkout master",
		"echo line3 > h",
		"git add h",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m qux --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repo interface {
			vcs.Merger
			ResolveRevision(spec string) (vcs.CommitID, error)
		}
		a, b string // can be any revspec; is resolved during the test

		wantMergeBase string // can be any revspec; is resolved during test
	}{
		"git cmd": {
			repo: makeGitRepositoryCmd(t, cmds...),
			a:    "master", b: "b2",
			wantMergeBase: "testbase",
		},
		"git go-git": {
			repo: makeGitRepositoryGoGit(t, cmds...),
			a:    "master", b: "b2",
			wantMergeBase: "testbase",
		},
	}

	for label, test := range tests {
		a, err := test.repo.ResolveRevision(test.a)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on a: %s", label, test.a, err)
			continue
		}

		b, err := test.repo.ResolveRevision(test.b)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on b: %s", label, test.b, err)
			continue
		}

		want, err := test.repo.ResolveRevision(test.wantMergeBase)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on wantMergeBase: %s", label, test.wantMergeBase, err)
			continue
		}

		mb, err := test.repo.MergeBase(a, b)
		if err != nil {
			t.Errorf("%s: MergeBase(%s, %s): %s", label, a, b, err)
			continue
		}

		if mb != want {
			t.Errorf("%s: MergeBase(%s, %s): got %q, want %q", label, a, b, mb, want)
			continue
		}
	}
}

func TestMerger_CrossRepoMergeBase(t *testing.T) {
	t.Parallel()

	// TODO(sqs): implement for hg
	// TODO(sqs): make a more complex test case

	cmdsA := []string{
		"echo line1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testbase",
	}
	cmdsB := []string{
		"echo line1 > f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git tag testbase",
		"git checkout -b b2",
		"echo line2 >> f",
		"git add f",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"git checkout master",
		"echo line3 > h",
		"git add h",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit -m qux --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
	}
	tests := map[string]struct {
		repoA interface {
			vcs.CrossRepoMerger
			ResolveRevision(spec string) (vcs.CommitID, error)
		}
		repoB vcs.Repository
		a, b  string // can be any revspec; is resolved during the test

		wantMergeBase string // can be any revspec; is resolved during test
	}{
		"git go-git": {
			repoA: makeGitRepositoryGoGit(t, cmdsA...),
			repoB: makeGitRepositoryGoGit(t, cmdsB...),

			a: "master", b: "b2",
			wantMergeBase: "testbase",
		},
		"git cmd": {
			repoA: makeGitRepositoryCmd(t, cmdsA...),
			repoB: makeGitRepositoryCmd(t, cmdsB...),

			a: "master", b: "b2",
			wantMergeBase: "testbase",
		},
	}

	for label, test := range tests {
		a, err := test.repoA.ResolveRevision(test.a)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on a: %s", label, test.a, err)
			continue
		}

		b, err := test.repoB.ResolveRevision(test.b)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on b: %s", label, test.b, err)
			continue
		}

		want, err := test.repoA.ResolveRevision(test.wantMergeBase)
		if err != nil {
			t.Errorf("%s: ResolveRevision(%q) on wantMergeBase: %s", label, test.wantMergeBase, err)
			continue
		}

		mb, err := test.repoA.CrossRepoMergeBase(a, test.repoB, b)
		if err != nil {
			t.Errorf("%s: CrossRepoMergeBase(%s, %s, %s): %s", label, a, test.repoB, b, err)
			continue
		}

		if mb != want {
			t.Errorf("%s: CrossRepoMergeBase(%s, %s, %s): got %q, want %q", label, a, test.repoB, b, mb, want)
			continue
		}
	}
}
