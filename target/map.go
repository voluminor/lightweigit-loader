// Code generated using '_generate/build_map'; DO NOT EDIT.
// Generation time: 2026-01-15T19:06:45Z

package target

// // // // // // // //

type ModType byte

const (
	ModTypeUnknown ModType = iota
	ModBitbucketTag
	ModBitbucketRelease
	ModGithubTag
	ModGithubRelease
	ModGitlabTag
	ModGitlabRelease
	ModGogsFamilyTag
	ModGogsFamilyRelease
)

func (m ModType) String() string {
	switch m {
	case ModBitbucketTag:
		return "BitbucketTag"
	case ModBitbucketRelease:
		return "BitbucketRelease"
	case ModGithubTag:
		return "GithubTag"
	case ModGithubRelease:
		return "GithubRelease"
	case ModGitlabTag:
		return "GitlabTag"
	case ModGitlabRelease:
		return "GitlabRelease"
	case ModGogsFamilyTag:
		return "GogsFamilyTag"
	case ModGogsFamilyRelease:
		return "GogsFamilyRelease"
	}

	return "unknown"
}
