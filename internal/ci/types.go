package ci

type CIHost string

const (
	HostUnknown CIHost = "unknown"
	HostLocal   CIHost = "local"
	HostGitHub  CIHost = "github"
	HostGitLab  CIHost = "gitlab"
)

// Context encapsulates everything needed to perform a reliable git diff.
type Context struct {
	Host CIHost

	// Event name (push, pull_request, merge_request_event, etc.)
	Event string

	// Refs to use for the diff
	BaseRef string // e.g., "origin/main", or a "before" SHA
	HeadRef string // e.g., "HEAD", or an "after" SHA

	// Raw SHAs (useful for push/force-push)
	BeforeSHA string
	AfterSHA  string

	// Flags
	IsPR        bool
	IsForcePush bool
	CommitCount int
}
