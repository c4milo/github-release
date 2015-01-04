# github-release

Yet another Github release command line tool. This one, though, is much more minimalistic and simple.

```
Github release tool.
Usage:
	github-release <user/repo> <tag> <branch> <description> <files>

<files> can be specified using glob patterns.
```

### Requirements
Make sure you have a Github token with the correct authorization scope to allow you to create releases. For more information about creating tokens please read https://help.github.com/articles/creating-an-access-token-for-command-line-use/

### Examples
Feel free to inspect this project's [Makefile](https://github.com/c4milo/github-release/blob/master/Makefile) for an example on how this tool can be used to create releases like this:

![](https://cldup.com/lTTZG_KQXI.png)
