package app

import "strings"

type PushPayload struct {
	Ref        string `json:"ref"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
}

func (p *PushPayload) RepoName() string { return p.Repository.Name }

func (p *PushPayload) Branch() string {
	if strings.HasPrefix(p.Ref, "refs/heads/") {
		return p.Ref[11:]
	}
	return ""
}
