# prometheus-gitlab-license-exporter

![ci](https://github.com/szevez/prometheus-gitlab-license-exporter/actions/workflows/ci.yml/badge.svg)

## Requirements

- GitLab API Token (to interact with license endpoints, you need to authenticate
yourself as an administrator, `read_api` is sufficient)
- GitLab URL

## Exported Metrics

| Name                              | Description                                                                          |
|-----------------------------------|--------------------------------------------------------------------------------------|
| gitlab_license_active_users       | Current active users that consume a license                                          |
| gitlab_license_expired            | Expiry status of the license                                                         |
| gitlab_license_expires_at         | Date the license expires at                                                          |
| gitlab_license_historical_max     | This is the highest peak of users on your installation since the license started     |
| gitlab_license_id                 | ID of the license                                                                    |
| gitlab_license_maximum_user_count | This is the highest peak of users on your installation since the license started     |
| gitlab_license_overage            | The difference between the number of billable users and the licensed number of users |
| gitlab_license_starts_at          | Date the license starts at                                                           |
| gitlab_license_user_limit         | The number of users the license is licensed for                                      |

## Docker

```shell
$ docker pull szevez/prometheus-gitlab-license-exporter:latest
$ docker run -p 9191:9191 \
    -e GITLAB_TOKEN=YOURTOKEN \
    -e GITLAB_URL=https://YOURURL \
    szevez/prometheus-gitlab-license-exporter:latest
```

## Local Development

- Requires go >= 1.18

```sh
$ export GITLAB_TOKEN=YOURTOKEN
$ export GITLAB_URL=https://YOURURL
$ go run main.go
```

Access exporter at `localhost:9191/metrics`
