package jj

const (
	RootChangeId = "zzzzzzzz"
)

type Commit struct {
	ChangeIdShort string
	ChangeId      string
	IsWorkingCopy bool
	Hidden        bool
	CommitIdShort string
	CommitId      string
}

func (c Commit) IsRoot() bool {
	return c.ChangeId == RootChangeId
}

func (c Commit) GetChangeId() string {
	if c.Hidden {
		return c.CommitId
	}
	return c.ChangeId
}
