package stash

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
)

type StashPullRequest struct {
	Project     StashProject           `json:"-"`
	Notifier    interface{}            `json:"-"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	State       string                 `json:"state"`
	Open        bool                   `json:"open"`
	Closed      bool                   `json:"closed"`
	FromRef     StashPullRequestRef    `json:"fromRef"`
	ToRef       StashPullRequestRef    `json:"toRef"`
	Locked      bool                   `json:"locked"`
	Reviewers   []StashPullRequestUser `json:"reviewers"`
}

type StashPullRequestRef struct {
	Id         string                     `json:"id"`
	Repository StashPullRequestRepository `json:"repository"`
}

type StashPullRequestRepository struct {
	Slug    string                  `json:"slug"`
	Project StashPullRequestProject `json:"project"`
}

type StashPullRequestUser struct {
	User map[string]string `json:"user"`
}

type StashPullRequestProject struct {
	Key string `json:"key"`
}

func (pr StashPullRequest) Create() error {
	log.Debug("Attempting to create a new pull request")

	pr.State = "OPEN"
	pr.Open = true
	pr.Closed = false

	requestUrl := fmt.Sprintf("/rest/api/1.0/projects/%s/repos/%s/pull-requests", pr.Project.Key, pr.Project.Repository)
	req := pr.Project.Request(requestUrl)
	req.Method = "POST"

	currentBranch, _ := pr.Project.CurrentBranch()

	repo := StashPullRequestRepository{Slug: pr.Project.Repository}
	project := StashPullRequestProject{Key: pr.Project.Key}

	pr.FromRef = StashPullRequestRef{}
	pr.FromRef.Id = fmt.Sprintf("refs/heads/%s", currentBranch)
	pr.FromRef.Repository = repo
	pr.FromRef.Repository.Project = project

	pr.ToRef = StashPullRequestRef{
		Id:         "refs/heads/master",
		Repository: repo,
	}
	pr.ToRef.Repository.Project = project

	for _, user := range pr.Project.Members {
		u := StashPullRequestUser{}
		u.User = make(map[string]string)
		u.User["name"] = user
		pr.Reviewers = append(pr.Reviewers, u)
	}

	req.Body = pr
	res, err := req.Do()

	if err != nil {
		log.Error(err.Error())
	}

	if res.Response.StatusCode != 201 {
		errors := StashAPIErrors{}
		res.Body.FromJsonTo(&errors)

		for _, v := range errors.Errors {
			log.Error(v["message"])
		}

		return nil
	}

	log.Debug(fmt.Sprintf("%+v", res.Response))
	//output, _ := json.MarshalIndent(pr, "", "    ")
	//os.Stdout.Write(output)
	return nil
}
