package dashboard

import (
	"context"
	"github.com/shuheiktgw/go-travis"
)

const (
	travisUrl    = "https://travis-ci.org/"
	TravisApiUrl = travis.ApiOrgUrl
)

type TravisInstance struct {
	Name   string
	Repos  []TravisRepository
	Client Travis
}

func (t *TravisInstance) Url() string {
	return travisUrl + t.Name
}

// Holds all dashboard relevant informations for a Travis instance
type TravisAggregation struct {
	Aggregation
	Jobs []TravisJob `json:"jobs"`
}

type Travis interface {
	Job(r TravisRepository) (TravisJob, error)
}

type TravisBuildStatus bool

type TravisRepository struct {
	Organization string
	Name         string
	Branch       string
}

type TravisClient struct {
	client *travis.Client
}

type TravisJob struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Color string `json:"color"`
}

func (j TravisJob) IsSuccessful() bool {
	return j.Color == "green" || j.Color == ""
}

func (c *TravisClient) status(r TravisRepository) (TravisBuildStatus, error) {
	branch, _, err := c.client.Branches.FindByRepoSlug(
		context.Background(),
		r.Organization+"/"+r.Name,
		r.Branch,
		&travis.BranchOption{},
	)

	if err != nil {
		return false, err
	}

	return *branch.LastBuild.State == "passed", nil
}

func (c *TravisClient) Job(r TravisRepository) (TravisJob, error) {
	status, err := c.status(r)
	job := TravisJob{
		Name: r.Name,
		URL:  travisUrl + r.Organization + "/" + r.Name,
	}

	color := "red"
	if err != nil {
		color = "grey"
	} else if status {
		color = "green"
	}

	job.Color = color
	return job, err
}

func NewTravisClient(baseUrl string, apiToken string) Travis {
	return &TravisClient{
		client: travis.NewClient(baseUrl, apiToken),
	}
}
