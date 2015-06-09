// Code generated by protoc-gen-gogo.
// source: vcs.proto
// DO NOT EDIT!

/*
Package vcs is a generated protocol buffer package.

It is generated from these files:
	vcs.proto

It has these top-level messages:
	Commit
	Signature
	Branch
	BehindAhead
	BranchesOptions
	Tag
	SearchOptions
	SearchResult
*/
package vcs

import proto "github.com/gogo/protobuf/proto"

// discarding unused import gogoproto "github.com/gogo/protobuf/gogoproto/gogo.pb"
import pbtypes "sourcegraph.com/sqs/pbtypes"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal

type Commit struct {
	ID        CommitID   `protobuf:"bytes,1,opt,name=id,proto3,customtype=CommitID" json:"id,omitempty"`
	Author    Signature  `protobuf:"bytes,2,opt,name=author" json:"author"`
	Committer *Signature `protobuf:"bytes,3,opt,name=committer" json:"committer,omitempty"`
	Message   string     `protobuf:"bytes,4,opt,name=message,proto3" json:"message,omitempty"`
	// Parents are the commit IDs of this commit's parent commits.
	Parents []CommitID `protobuf:"bytes,5,rep,name=parents,customtype=CommitID" json:"parents,omitempty"`
}

func (m *Commit) Reset()         { *m = Commit{} }
func (m *Commit) String() string { return proto.CompactTextString(m) }
func (*Commit) ProtoMessage()    {}

func (m *Commit) GetAuthor() Signature {
	if m != nil {
		return m.Author
	}
	return Signature{}
}

func (m *Commit) GetCommitter() *Signature {
	if m != nil {
		return m.Committer
	}
	return nil
}

type Signature struct {
	Name  string            `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Email string            `protobuf:"bytes,2,opt,name=email,proto3" json:"email,omitempty"`
	Date  pbtypes.Timestamp `protobuf:"bytes,3,opt,name=date" json:"date"`
}

func (m *Signature) Reset()         { *m = Signature{} }
func (m *Signature) String() string { return proto.CompactTextString(m) }
func (*Signature) ProtoMessage()    {}

func (m *Signature) GetDate() pbtypes.Timestamp {
	if m != nil {
		return m.Date
	}
	return pbtypes.Timestamp{}
}

// A Branch is a VCS branch.
type Branch struct {
	// Name is the name of this branch.
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	// Head is the commit ID of this branch's head commit.
	Head CommitID `protobuf:"bytes,2,opt,name=head,proto3,customtype=CommitID" json:"head,omitempty"`
	// Commit optionally contains commit information for this branch's head commit.
	// It is populated if IncludeCommit option is set.
	Commit *Commit `protobuf:"bytes,4,opt,name=commit" json:"commit,omitempty"`
	// Counts optionally contains the commit counts relative to specified branch.
	Counts *BehindAhead `protobuf:"bytes,3,opt,name=counts" json:"counts,omitempty"`
}

func (m *Branch) Reset()         { *m = Branch{} }
func (m *Branch) String() string { return proto.CompactTextString(m) }
func (*Branch) ProtoMessage()    {}

func (m *Branch) GetCommit() *Commit {
	if m != nil {
		return m.Commit
	}
	return nil
}

func (m *Branch) GetCounts() *BehindAhead {
	if m != nil {
		return m.Counts
	}
	return nil
}

// BehindAhead is a set of behind/ahead counts.
type BehindAhead struct {
	Behind uint32 `protobuf:"varint,1,opt,name=behind,proto3" json:"behind,omitempty"`
	Ahead  uint32 `protobuf:"varint,2,opt,name=ahead,proto3" json:"ahead,omitempty"`
}

func (m *BehindAhead) Reset()         { *m = BehindAhead{} }
func (m *BehindAhead) String() string { return proto.CompactTextString(m) }
func (*BehindAhead) ProtoMessage()    {}

// BranchesOptions specifies options for the list of branches returned by
// (Repository).Branches.
type BranchesOptions struct {
	// IncludeCommit controls whether complete commit information is included.
	IncludeCommit bool `protobuf:"varint,2,opt,name=include_commit,proto3" json:"include_commit,omitempty" url:",omitempty"`
	// BehindAheadBranch specifies a branch name. If set to something other than blank
	// string, then each returned branch will include a behind/ahead commit counts
	// information against the specified base branch. If left blank, then branches will
	// not include that information and their Counts will be nil.
	BehindAheadBranch string `protobuf:"bytes,1,opt,name=behind_ahead_branch,proto3" json:"behind_ahead_branch,omitempty" url:",omitempty"`
}

func (m *BranchesOptions) Reset()         { *m = BranchesOptions{} }
func (m *BranchesOptions) String() string { return proto.CompactTextString(m) }
func (*BranchesOptions) ProtoMessage()    {}

// A Tag is a VCS tag.
type Tag struct {
	Name     string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	CommitID CommitID `protobuf:"bytes,2,opt,name=commit_id,proto3,customtype=CommitID" json:"commit_id,omitempty"`
}

func (m *Tag) Reset()         { *m = Tag{} }
func (m *Tag) String() string { return proto.CompactTextString(m) }
func (*Tag) ProtoMessage()    {}

// SearchOptions specifies options for a repository search.
type SearchOptions struct {
	// the query string
	Query string `protobuf:"bytes,1,opt,name=query,proto3" json:"query,omitempty"`
	// currently only FixedQuery ("fixed") is supported
	QueryType string `protobuf:"bytes,2,opt,name=query_type,proto3" json:"query_type,omitempty"`
	// the number of lines before and after each hit to display
	ContextLines int32 `protobuf:"varint,3,opt,name=context_lines,proto3" json:"context_lines,omitempty"`
	// max number of matches to return
	N int32 `protobuf:"varint,4,opt,name=n,proto3" json:"n,omitempty"`
	// starting offset for matches (use with N for pagination)
	Offset int32 `protobuf:"varint,5,opt,name=offset,proto3" json:"offset,omitempty"`
}

func (m *SearchOptions) Reset()         { *m = SearchOptions{} }
func (m *SearchOptions) String() string { return proto.CompactTextString(m) }
func (*SearchOptions) ProtoMessage()    {}

// A SearchResult is a match returned by a search.
type SearchResult struct {
	// File is the file that contains this match.
	File string `protobuf:"bytes,1,opt,name=file,proto3" json:"file,omitempty"`
	// The byte range [start,end) of the match.
	StartByte uint32 `protobuf:"varint,2,opt,name=start_byte,proto3" json:"start_byte,omitempty"`
	EndByte   uint32 `protobuf:"varint,3,opt,name=end_byte,proto3" json:"end_byte,omitempty"`
	// The line range [start,end] of the match.
	StartLine uint32 `protobuf:"varint,4,opt,name=start_line,proto3" json:"start_line,omitempty"`
	EndLine   uint32 `protobuf:"varint,5,opt,name=end_line,proto3" json:"end_line,omitempty"`
	// Match is the matching portion of the file from [StartByte,
	// EndByte).
	Match []byte `protobuf:"bytes,6,opt,name=match,proto3" json:"match,omitempty"`
}

func (m *SearchResult) Reset()         { *m = SearchResult{} }
func (m *SearchResult) String() string { return proto.CompactTextString(m) }
func (*SearchResult) ProtoMessage()    {}

func init() {
}