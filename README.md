# github-release

Yet another Github release command line tool. This one, though, is much more minimalistic and simple to use.

```
Github command line release tool.

Usage:
	github-release <user/repo> <tag> <branch> <description> <files>

Parameters:
	<user/repo>: Github user and repository
	<tag>: Used to created the release. It is also used as the release's name
	<branch>: Reference from where to create the provided <tag>, if it does not exist
	<description>: The release description
	<files>: Glob pattern describing the list of files to include in the release

Options:
	-version: Displays version

Environment variables:
  DEBUG: Allows you to run github-release in debugging mode. DO NOT do this if you are attempting to upload big files.
  GITHUB_TOKEN: Must be set in order to interact with Github's API
  GITHUB_USER: Just in case you want an alternative way of providing your github user
  GITHUB_REPO: Just in case you want an alternative way of providing your github repo
  GITHUB_API: Github API endpoint. Set to https://api.github.com/repos/:github-user/:github-repo" by default

Before using this tool make sure you set the environment variable GITHUB_TOKEN
with a valid Github token and correct authorization scopes to allow you to create releases
in your project. For more information about creating Github tokens please read the
official Github documentation at https://help.github.com/articles/creating-an-access-token-for-command-line-use/

Author: https://github.com/c4milo
License: http://mozilla.org/MPL/2.0/

```

### Examples
Feel free to inspect this project's [Makefile](https://github.com/c4milo/github-release/blob/master/Makefile) for an example on how this tool can be used to create releases like this:

![](https://cldup.com/6Slplyys6X.png)

 
