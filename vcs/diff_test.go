			repo: makeGitRepositoryCmd(t, cmds...),
			repo: makeHgRepositoryCmd(t, hgCommands...),
			baseRepo: makeGitRepositoryCmd(t, cmds...),
			headRepo: makeGitRepositoryCmd(t, cmds...),