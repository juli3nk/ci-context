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
// If since is provided, it acts as a floor: the returned BaseRef will never be
// older than since. If since is not an ancestor of head, an error is returned.
func GetRefs(base, since string) (*Refs, error) {
	// Manual override
	if base != "" {
		baseRef := resolveBase(base)
		headRef := "HEAD"
		baseRef, err := clampBase(baseRef, headRef, since)
		if err != nil {
			return nil, err
		}
		return &Refs{
			BaseRef:     baseRef,
			HeadRef:     headRef,
			CommitCount: countRefs(baseRef, headRef),
		}, nil
	}

	ci := Detect()

	baseRef := resolveBase(ci.BaseRef)
	headRef := ci.HeadRef
	if headRef == "" {
		headRef = "HEAD"
	}

	switch ci.Host {
	case HostGitHub:
		if ci.IsPR || ci.IsForcePush {
			baseRef, err := clampBase(baseRef, headRef, since)
			if err != nil {
				return nil, err
			}
			commitCount := max(ci.CommitCount, countRefs(baseRef, headRef))
			return &Refs{BaseRef: baseRef, HeadRef: headRef, CommitCount: commitCount}, nil
		}
		baseRef = resolveBase(ci.BeforeSHA)
		baseRef, err := clampBase(baseRef, ci.AfterSHA, since)
		if err != nil {
			return nil, err
		}
		commitCount := max(ci.CommitCount, countRefs(baseRef, ci.AfterSHA))
		return &Refs{BaseRef: baseRef, HeadRef: ci.AfterSHA, CommitCount: commitCount}, nil
	case HostGitLab:
		if ci.IsPR {
			baseRef, err := clampBase(baseRef, headRef, since)
			if err != nil {
				return nil, err
			}
			commitCount := max(ci.CommitCount, countRefs(baseRef, headRef))
			return &Refs{BaseRef: baseRef, HeadRef: headRef, CommitCount: commitCount}, nil
		}
		baseRef = resolveBase(ci.BeforeSHA)
		baseRef, err := clampBase(baseRef, ci.AfterSHA, since)
		if err != nil {
			return nil, err
		}
		commitCount := max(ci.CommitCount, countRefs(baseRef, ci.AfterSHA))
		return &Refs{BaseRef: baseRef, HeadRef: ci.AfterSHA, CommitCount: commitCount}, nil
	default:
		baseRef, err := clampBase(baseRef, headRef, since)
		if err != nil {
			return nil, err
		}
		commitCount := max(ci.CommitCount, countRefs(baseRef, headRef))
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
