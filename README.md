# change-aws-credentials

A simple tool to change your AWS credentials quickly from the commandline.

Currently supports changing your password, will eventually support changing your MFA and Access/Secret Keys

# Prerequisites

Your AWS users must have the ability to change their own IAM credentials. You can see how to configure this [here](http://docs.aws.amazon.com/IAM/latest/UserGuide/tutorial_users-self-manage-mfa-and-creds.html)

# Installation

### OS X

Install via homebrew: `brew install jaxxstorm/tap/change-aws-credentials`

### Linux

Install the Go binary and place it in your `$PATH`. You can download the latest version using this handy one-liner:

```
curl -s https://api.github.com/repos/jaxxstorm/change-aws-credentials/releases/latest | jq -r '.assets[]| select(.browser_download_url | contains("linux")) | .browser_download_url' | wget -i -
```

### Windows

Download the latest release from the [releases](https://github.com/jaxxstorm/change-aws-credentials/releases/latest) page and install as appropriate.

# Usage

```bash
Allows users to quickly reset their AWS credentials without
having to burden an administrator

Usage:
  change-aws-credentials [flags]
  change-aws-credentials [command]

Available Commands:
  help        Help about any command
  keys        Rotate your AWS keys
  password    Change your AWS Password

Flags:
  -P, --awsprofile string   AWS Profile to Change Credentials for
      --config string       config file (default is $HOME/.change-aws-password.yaml)
  -h, --help                help for change-aws-credentials

Use "change-aws-credentials [command] --help" for more information about a command.
```

## Profiles

You can specify the AWS profile you want to change the password for using the `-P` flag. If you don't explicitly specify this, the tool will warn you which profile you're changing using by checking the `AWS_PROFILE` environment variable

# Building

If you want to contribute, we use glide for dependency management, so it should be as simple as:

 - cloning this repo into `$GOPATH/src/github.com/jaxxstorm/change-aws-credentials 
 - run glide install from the directory 
 - run go build -o change-aws-credentials main.go




