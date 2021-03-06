# aws-lambda-ddos-hangman [![CircleCI](https://circleci.com/gh/moznion/aws-lambda-ddos-hangman.svg?style=svg)](https://circleci.com/gh/moznion/aws-lambda-ddos-hangman) [![GoDoc](https://godoc.org/github.com/moznion/aws-lambda-ddos-hangman?status.svg)](https://godoc.org/github.com/moznion/aws-lambda-ddos-hangman) [![Go Report Card](https://goreportcard.com/badge/github.com/moznion/aws-lambda-ddos-hangman)](https://goreportcard.com/report/github.com/moznion/aws-lambda-ddos-hangman)

ddos-hangman is an AWS Lambda function to drop the inbound requests from attackers on a network ACL layer.

## Flow diagram

```
[clinet] --put/delete item--> [dynamodb] --insert event--> +----------------+ --apply denying--> [NACL (Network ACL Entry)]
                                  |                        | hangman lambda |                                ^
                                  +--------remove event--> +----------------+ --release denying--------------+
```

## Description

This lambda function aims to drop the inbound requests from attackers.

Once a client has wanted to deny access from the attacker, it just only needs to insert an item into DynamoDB.
Then the lambda function gets the inserted item through the DynamoDB Streams and applies the configuration
to deny requests from attacker according to the item to Network ACL.
(NOTICE: of course the DynamoDB table has to be activated the DynamoDB Streams)

If the number of entries of network ACL reaches the maximum limit, it removes the oldest NACL entry like FIFO.

After it, if a client would like to release denying configuration, it needs to remove the item from the DynamoDB table.
This lambda function releases the configuration of Network ACL as similar to inserting.
(NOTE: this method is much effective when used along with DynamoDB's TTL mechanism)

## Pre requirements

### DynamoDB

- make a table to use for this purpose
- enable the DynamoDB Streams for the table
- optional: apply TTL option for an attribute of the table

### Network ACL

- create a network ACL to use for this purpose

### Deployment toolchain

- install following software
  - go 1.13 or later
  - AWS SAM CLI

## How to deploy

At first, please edit a `template.yaml` file as you like, after:

```
$ make deploy S3BUCKET="your-s3-backet-to-store-Cfn-stack"
```

It deploys the lambda function by AWS SAM.

## Configurable environment variables for the lambda

- `REGION`:
  - AWS region name which you would like to use this on
- `TABLE_NAME`
  - DynamoDB table name for denied applicants
- `BEGIN_RULE_NUMBER`:
  - the beginning number of ACL rule; this lambda tries to create ACL rule according to this value incrementally
    - e.g. if you specify this parameter as 100, this lambda tries to create an ACL rule with #100. If the number has been already used, it retries to create that with #101...
- `IGNORE_ERROR`:
  - if this parameter is __not__ empty, this function ignores (but logs) the errors
  - DynamoDB Streams' shard iterator is stacked if the function's error has not been cleared
    - so perhaps this parameter is necessary to proceed forward continuously

## What an item should be put on the DynamoDB table

- `subject` (string);
  - the partition key
  - this function applies denying rule for this target
  - it has to follow the following regexp pattern:
    - `(?<created_at_epoch_millis>[0-9]+):(?<cidr>[0-9]{1,3}[.][0-9]{1,3}[.][0-9]{1,3}[.][0-9]{1,3}/[0-9]{1,2}):(?<protocol>[0-9]+)(?::(?<port_range>[0-9]{1,5}-[0-9]{1,5}))?`
      - `protocol` means the protocol number to deny (e.g. 6=TCP, 17=UDP)
    - examples:
      - `1582425243392:192.0.2.1/32:6`
      - `1582425243392:192.0.2.1/32:6:22-80`
  - if ports were specified, the ACL will be applied for only the ports
    - on the other hand, the ACL will be applied all of the ports
  - notice: the port specifying is only available on TCP/UDP
- `networkAclId` (string):
  - network ACL ID to apply the rules

You can use the library that is provided by this repository for this purpose if you use golang: [GoDoc](https://godoc.org/github.com/moznion/aws-lambda-ddos-hangman/execute/repo#DeniedApplicantRepo)

### Example

```json
{
  "subject": {
    "S": "1582425243392:192.0.2.1/32:6"
  },
  "networkAclID": {
    "S": "acl-0ea1f54ca7EXAMPLE"
  }
}
```

## License

MIT

