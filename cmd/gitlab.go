package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/xanzy/go-gitlab"

	"github.com/ducmeit1/gitlab-runner-clean/logger"
)

type runner struct {
	Id   int
	Name string
}

type stat struct {
	total   int
	deleted int
	error   int
}

type GitLabClient struct {
	context context.Context
	client  *gitlab.Client
	runner  chan runner
	stat    *stat
	done    chan bool
}

func NewGitLabClient(ctx context.Context) (*GitLabClient, error) {
	address := os.Getenv("GITLAB_ADDRESS")
	if address == "" {
		return nil, fmt.Errorf("gitlab address must not be empty")
	}

	token := os.Getenv("GITLAB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("gitlab token must not be empty")
	}

	logger.Infof("ADDRESS: %v, TOKEN: %v", address, token)

	client, err := gitlab.NewClient(token, gitlab.WithBaseURL(address))
	if err != nil {
		return nil, err
	}

	return &GitLabClient{
		context: ctx,
		client:  client,
		runner:  make(chan runner),
		stat: &stat{
			total:   0,
			deleted: 0,
			error:   0,
		},
		done: make(chan bool),
	}, nil
}

func (_this *GitLabClient) GetAllOfflineRunners() (int, error) {
	runners, response, err := _this.client.Runners.ListAllRunners(&gitlab.ListRunnersOptions{
		Status: gitlab.String("offline"),
	})

	if err != nil {
		return 0, fmt.Errorf("fetch offline runners has error: %v", err.Error())
	}

	if response.StatusCode != 200 {
		return 0, fmt.Errorf("gitlab api return status code != 200")
	}

	total := len(runners)
	_this.stat.total = total

	for _, r := range runners {
		_this.runner <- runner{
			Id:   r.ID,
			Name: r.Name,
		}
	}

	return total, nil
}

func (_this *GitLabClient) DeleteAllOfflineRunners() {
	for {
		select {
		case <-_this.context.Done():
			return
		case <-_this.done:
			return
		case r := <-_this.runner:
			response, err := _this.client.Runners.DeleteRegisteredRunnerByID(r.Id)
			if err != nil {
				_this.stat.error++
				logger.Errorf("delete runner: %v (%v) has error: %v", r.Id, r.Name, err.Error())
			} else {
				if response.StatusCode != 200 {
					_this.stat.error++
					logger.Warnf("delete runner: %v (%v) status != 200: %v", r.Id, r.Name, response.StatusCode)
				} else {
					_this.stat.deleted++
					logger.Infof("delete runner: %v (%v) successful", r.Id, r.Name)
				}
			}

			if _this.stat.deleted+_this.stat.error == _this.stat.total {
				_this.done <- true
			}
		}
	}
}
