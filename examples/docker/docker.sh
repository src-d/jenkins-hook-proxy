#!/usr/bin/env bash

docker run --rm \
-e LOG_LEVEL="${LOG_LEVEL}" \
-e JENKINS_ENDPOINT="${JENKINS_ENDPOINT}" \
-e SLACK_TOKEN="${SLACK_TOKEN}" \
-e SLACK_CHANNEL="${SLACK_CHANNEL}" \
-e GITHUB_SECRET="${GITHUB_SECRET}" \
-p 9000:9000 \
--name proxy srcd/jenkins-hook-proxy:latest
