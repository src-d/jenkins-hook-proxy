# jenkins-hook-proxy

## Logic
**jenkins-hook-proxy** is a server, which handler does the next steps:
1) receive github webhook [push event](https://developer.github.com/v3/activity/events/types/#pushevent)
2) parses retrieved data using github webhook [secret](https://developer.github.com/webhooks/securing/)
3) composes request to trigger Jenkins [generic webhook](https://wiki.jenkins.io/display/JENKINS/Generic+Webhook+Trigger+Plugin) endpoint by setting a special `token` as a URL value, that is equal to the repository name.
4) log results to Stdout and slack channel(if configured)

## Role in cluster
**jenkins-hook-proxy** should be:
1) located inside the cluster
2) exposed to github, but protected with github webhook [secret](https://developer.github.com/webhooks/securing/)
3) able to reach jenkins to trigger the jobs 

## Requirements
- `JENKINS_ENDPOINT`
- `GITHUB_SECRET`
- `SLACK_TOKEN` (optional)
- `SLACK_CHANNEL` (optional)

## Usage
See examples directory
