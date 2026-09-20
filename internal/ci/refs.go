package ci

import (
	"encoding/json"
	"fmt"
)

type Refs struct {
	BaseRef     string `json:"base_ref"`
	HeadRef     string `json:"head_ref"`
	CommitCount int    `json:"commit_count"`
}

// GetRefs detects the scope of the change.
func GetRefs(base string) (*Refs, error) {
	// Manual override
	if base != "" {
		return &Refs{
			BaseRef:     resolveBase(base),
			HeadRef:     "HEAD",
			CommitCount: countRefs(resolveBase(base), "HEAD"),
		}, nil
	}

	ci := Detect()

	baseRef := resolveBase(ci.BaseRef)
	headRef := ci.HeadRef
	if headRef == "" {
		headRef = "HEAD"
	}
	commitCount := max(ci.CommitCount, countRefs(baseRef, headRef))

	switch ci.Host {
	case HostGitHub:
		if ci.IsPR {
			return &Refs{BaseRef: baseRef, HeadRef: headRef, CommitCount: commitCount}, nil
		} else if ci.IsForcePush {
			return &Refs{BaseRef: baseRef, HeadRef: headRef, CommitCount: commitCount}, nil
		} else {
			return &Refs{BaseRef: resolveBase(ci.BeforeSHA), HeadRef: ci.AfterSHA, CommitCount: max(ci.CommitCount, countRefs(resolveBase(ci.BeforeSHA), ci.AfterSHA))}, nil
		}
	case HostGitLab:
		if ci.IsPR {
			return &Refs{BaseRef: baseRef, HeadRef: headRef, CommitCount: commitCount}, nil
		} else {
			return &Refs{BaseRef: resolveBase(ci.BeforeSHA), HeadRef: ci.AfterSHA, CommitCount: max(ci.CommitCount, countRefs(resolveBase(ci.BeforeSHA), ci.AfterSHA))}, nil
		}
	default:
		return &Refs{BaseRef: baseRef, HeadRef: headRef, CommitCount: commitCount}, nil
	}
}

func (r *Refs) GithubOutput() string {
	return fmt.Sprintf("base_ref=%s\nhead_ref=%s\ncommit_count=%d", r.BaseRef, r.HeadRef, r.CommitCount)
}

func (r *Refs) Json() string {
	b, err := json.Marshal(r)
	if err != nil {
		return ""
	}

	return string(b)
}
