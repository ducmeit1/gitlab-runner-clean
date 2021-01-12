# Clean Offline Gitlab Runners Job

The job requires you have to input the address of gitlab server and API token.

## Local Run

- Export Environment variable: `GITLAB_ADDRESS` and `GITLAB_TOKEN`
- Run with Go

```go
go run main.go
```

# Kubernetes CronJob

- Change values in Secret Template
- Run with kubectl

```shell
kubectl apply -f kubernetes.yml
```

