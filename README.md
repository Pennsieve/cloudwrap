# Cloudwrap

[![CI](https://github.com/Pennsieve/cloudwrap/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/Pennsieve/cloudwrap/actions/workflows/ci.yml)

Cloudwrap is an opinionated utility for fetching configuration from
**AWS SSM Parameter Store**. Its primary use is to wrap a command: the executed
command is injected with the configuration as environment variables.

Parameters are fetched by resource path. A path has exactly three components —
a key nested under an environment and a service name:

```
/{environment}/{service}/key
```

Multiple services can be provided (comma-separated). Their configurations are
merged; when two services define the same key, the **left-most** service in the
list wins.

Key names are converted from kebab-case to upper-snake-case: the final path
segment is uppercased and `-` is replaced with `_` (e.g. `one-key` → `ONE_KEY`).

> This tool only *fetches* configuration. Use another mechanism of your choice to
> *set* values in SSM Parameter Store.

## Install

Download a binary, `.deb`, or `.apk` from the
[Releases](https://github.com/Pennsieve/cloudwrap/releases) page, or build from
source:

```
go install github.com/pennsieve/cloudwrap@latest
```

## Usage

The environment and service are given as flags (or via the
`CLOUDWRAP_ENVIRONMENT` / `CLOUDWRAP_SERVICE` environment variables). Cloudwrap
uses the standard AWS credential chain and the `us-east-1` region.

Describe keys for a service (no values):

```
$ cloudwrap -e staging -s service-name-test describe
+---------+---------+----------------------+---------------------+
|   KEY   | VERSION |  LAST_MODIFIED_USER  | LAST_MODIFIED_DATE  |
+---------+---------+----------------------+---------------------+
| one-key |       1 | vienna@cloudwrap.com | 2018-04-24 19:36:02 |
| two     |       1 | lachy@cloudwrap.com  | 2018-04-24 19:36:16 |
+---------+---------+----------------------+---------------------+
```

Print key/value pairs to stdout:

```
$ cloudwrap -e staging -s service-name-test stdout
ONE_KEY=valueone
TWO=valuetwo
```

Write key/value pairs to a file as `export` statements (the parent directory
must already exist):

```
$ cloudwrap -e staging -s service-name-test file ./config.env
$ cat ./config.env
export ONE_KEY=valueone
export TWO=valuetwo
```

Execute a command with the configuration injected as environment variables. The
wrapped command's exit code is forwarded, and signals (Ctrl-C) are passed
through to it:

```
$ cloudwrap -e staging -s service-name-test exec env
ONE_KEY=valueone
TWO=valuetwo
...
```

Merge multiple services (left-most wins on conflict):

```
$ cloudwrap -e staging -s base-service,auth-service stdout
```

See `cloudwrap --help` and `cloudwrap <command> --help` for full details.

## Permissions

### AWS IAM

Minimal IAM policy needed to wrap a program that reads from SSM Parameter Store.
The `kms:Decrypt` permission is only needed if your parameters are `SecureString`
values; adjust the KMS key/alias if you did not use the SSM default.

#### Command

```
cloudwrap -e dev -s auth-service exec java -jar {jar-name}.jar
```

#### Resource

```hcl
resource "aws_iam_role_policy" "parameters" {
  name = "dev-auth-service-parameter-policy"
  role = "${var.role_id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ssm:GetParameter",
        "ssm:GetParameters",
        "ssm:GetParametersByPath",
        "ssm:DescribeParameters"
      ],
      "Effect": "Allow",
      "Resource": [
        "arn:aws:ssm:${var.aws_region}:${var.aws_account_id}:parameter/dev/auth-service/*"
      ]
    },
    {
      "Action": [
        "kms:Decrypt"
      ],
      "Effect": "Allow",
      "Resource": "arn:aws:kms:${var.aws_region}:${var.aws_account_id}:key/alias/aws/ssm"
    }
  ]
}
EOF
}
```

## Development

```
go build ./...     # build
go test ./...      # run tests
go vet ./...       # static checks
```

Releases are cut by pushing a semver tag (e.g. `1.0.0`); CI runs
[GoReleaser](https://goreleaser.com/) to build binaries and `.deb`/`.apk`
packages and publish them to GitHub Releases.

## License

This project is licensed under the Apache License, Version 2.0
([LICENSE-APACHE](LICENSE-APACHE) or http://www.apache.org/licenses/LICENSE-2.0).
