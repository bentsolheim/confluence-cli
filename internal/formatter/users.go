package formatter

import (
	"regexp"
)

var userRefRe = regexp.MustCompile(`<ri:user\b[^>]*>`)
var userAttrRe = regexp.MustCompile(`ri:(userkey|username)\s*=\s*"([^"]+)"`)

// UserRefs holds the unique user keys and usernames mentioned in storage format.
type UserRefs struct {
	Keys      []string
	Usernames []string
}

// ExtractUserRefs finds the unique ri:userkey and ri:username values of
// <ri:user> elements in Confluence storage format, in order of appearance.
func ExtractUserRefs(storageHTML string) UserRefs {
	var refs UserRefs
	seen := map[string]bool{}
	for _, tag := range userRefRe.FindAllString(storageHTML, -1) {
		for _, m := range userAttrRe.FindAllStringSubmatch(tag, -1) {
			kind, value := m[1], m[2]
			if seen[kind+":"+value] {
				continue
			}
			seen[kind+":"+value] = true
			if kind == "userkey" {
				refs.Keys = append(refs.Keys, value)
			} else {
				refs.Usernames = append(refs.Usernames, value)
			}
		}
	}
	return refs
}

// UserLabel returns the label used for a resolved user mention,
// e.g. "Solheim, Bent Andre / k77319".
func UserLabel(displayName, username string) string {
	switch {
	case displayName != "" && username != "":
		return displayName + " / " + username
	case displayName != "":
		return displayName
	default:
		return username
	}
}
