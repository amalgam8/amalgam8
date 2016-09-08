# Contributing

## Getting help
If you are looking for something to work on, or need some assistance in debugging a problem or working out a fix to an issue, please visit our [Slack channels](https://amalgam8.slack.com/) and will gladly help.
To join Amalgam8 Slack, please use [auto-invite](https://amalgam8-slack-invite.mybluemix.net/). 

## Reporting bugs
If you are a user and you find a bug, please submit an issue to the appropriate sub-project. 
If unsure, please consult with us on the [Slack users channel](https://amalgam8.slack.com/messages/users/). 
Please try to provide sufficient information for someone else to reproduce the issue. 
One of the project's maintainers should respond to your issue within 24 hours. 
If not, please bump the issue and request that it be reviewed.

## Submitting fixes and enhancements
Review the issues list for a sub-project and find something that interests you. 
Usually there will be a comment in the issue that indicates whether someone is already assigned to the issue. 
If no one has already taken it, then add a comment assigning the issue to yourself. 
You could also join the conversation on the [Slack developer channel](https://amalgam8.slack.com/messages/devel/).

We are using [GitHub Forking](https://guides.github.com/activities/forking/) process to manage code contributions. 
If you are unfamiliar, please review that link (or [this one](https://www.atlassian.com/git/tutorials/comparing-workflows/forking-workflow)) before proceeding.

When you need feedback or help, or you think the branch is ready for merging, open a pull request.
Please make sure you have first successfully built and tested the code beforehand.
  
### Coding guidelines

#### Coding in Go
For sub-projects using Go&trade; please follow the [best practices](http://golang.org/doc/effective_go.html).
You must use the default Go formatting guidelines for your source code and
run static analysis tools before submitting a pull request. These tasks can
be run in an automated fashion by invoking
```bash
make precommit
```
You can install a git-hook into the local `.git/hooks/` directory, as a
pre-commit ot pre-push hook.

<!-- and run the following tools against your Go code and fix all errors and warnings: -->
<!-- - [golint](https://github.com/golang/lint) -->
<!-- - [go vet](https://golang.org/cmd/vet/) -->
<!-- - [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) -->

#### Updating dependencies

Whenver adding/removing/updating dependencies (`make depend.update`), make sure to do so in a single, isolated commit, that touches only `glide.*` and `vendor/*` files.  
This avoids polluting/obfuscating other commits in the same pull request, and keeps it easy to review.
  
## Legal stuff

**Note:** Each source file must include a license header for the Apache Software License 2.0. 
A template of that header can be found [here](http://www.apache.org/licenses/LICENSE-2.0#apply).

We have tried to make it as easy as possible to make contributions. 
This applies to how we handle the legal aspects of contribution. 
Contributions require sign-off. 
The sign-off is a simple line at the end of the commit message for the patch or pull request.
It certifies that you wrote it or otherwise have the right to pass it on as an open-source patch. 

If you can certify the [Developer's Certificate of Origin 1.1 (DCO)](http://elinux.org/Developer_Certificate_Of_Origin) then you just add a `Signed-off-by` line, which indicates that the you accept the DCO

    Signed-off-by: Jane Doe <jane.doe@domain.com>

When committing using the command line you can sign off using the `--signoff` or `-s` flag. 
This adds a Signed-off-by line by the committer at the end of the commit log message.

    git commit -s -m "Commit message"
