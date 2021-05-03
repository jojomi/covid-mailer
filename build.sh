#!/bin/sh
set -ex

GIT_COMMIT=$(git rev-list -1 HEAD)
GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
GIT_VERSION=$(git rev-list --count HEAD)
GIT_DATE=$(git show -s --format=%ci HEAD)
GIT_STATE=$(git diff --quiet && echo 'clean' || echo 'dirty')
GIT_REMOTE=$(git config --get remote.origin.url)

go generate

# trying to be reproducible
CGO_ENABLED=0 go install -trimpath -ldflags "-buildid= -X 'main.GitCommit=$GIT_COMMIT' -X 'main.GitBranch=$GIT_BRANCH' -X 'main.GitDate=$GIT_DATE' -X 'main.GitVersion=$GIT_VERSION' -X 'main.GitState=$GIT_STATE'  -X 'main.GitRemote=$GIT_REMOTE'"
