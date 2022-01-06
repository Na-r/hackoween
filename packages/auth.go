package packages

import (
	_ "golang.org/x/oauth2/github"
	_ "golang.org/x/oauth2/gitlab"
	_ "golang.org/x/oauth2/google"
)

func Test() int {
	return 6
}
