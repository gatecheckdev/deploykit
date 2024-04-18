# GitOps-Toolkit

A simple utility for performing common GitOps tasks.

If you are using ssh to clone, you may need to add to the `known_host` since
this isn't done by GTK.

```shell
ssh-keyscan github.com >> ~/.ssh/known_hosts
```

## Use Cases

### GitOps Style Deployment in a CI/CD Pipeline

This tool solves the issue of having to write the logic necessary for a
rebase / push loop when running a deployment pipeline in a CI/CD pipeline.

For example, if you have a single pipeline that updates multiple services in a
kustomize manifest repository, it's possible to create a race condition if
both services attempt to push a commit to the repository at the same time.

One solution is to run the deployments in sequence, where pipeline A waits for
pipeline B to finish running before starting.
Depending on the structure of the pipeline and CI/CD platform of choice, this
may or may not be possible.

An alternative solution is to rebase and retry if the push fails.
The logic here seems simple but usually results in unreadable bash script that's
hard to debug and even harder to maintain.

GTK handles this complexity by providing two methods, exponential back-off and
random back-off.

It also handles cloning the manifest repository and running the kustomize set
image command.

```shell
gtk deploy kustomize --repository [SOME REPO URL] --service [SOME SERVICE] --image some-image:latest --service-directory [SOME SUB DIR (ex. prod/dev)]
```
