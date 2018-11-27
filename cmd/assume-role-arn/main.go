package main

import (
	"flag"
	"fmt"
	"os"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

var roleARN string
var roleName string
var externalID string

func init() {
	flag.StringVar(&roleARN, "role", "", "role arn")
	flag.StringVar(&roleARN, "r", "", "role arn (shorthand)")

	flag.StringVar(&roleName, "name", "assumed-role", "role session name")
	flag.StringVar(&roleName, "n", "assumed-role", "role session name (shorthand)")

	flag.StringVar(&externalID, "extid", "", "external id")
	flag.StringVar(&externalID, "e", "", "external id (shorthand)")

	flag.Parse()

	if roleARN == "" {
		panic("Role ARN cannot be empty")
	}
}

func prepareAssumeInput() *sts.AssumeRoleInput {
	input := &sts.AssumeRoleInput{
		RoleArn:         aws.String(roleARN),
		RoleSessionName: aws.String(roleName),
	}

	if externalID != "" {
		input.ExternalId = aws.String(externalID)
	}

	return input
}

func getSession() *session.Session {
	region := "us-east-1"
	sess, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Region: aws.String(region),
		},
	})

	if err != nil {
		panic(err)
	}

	return sess
}

func assumeRole(sess *session.Session, input *sts.AssumeRoleInput) *sts.AssumeRoleOutput {
	svc := sts.New(sess)
	out, err := svc.AssumeRole(input)
	if err != nil {
		panic(err)
	}
	return out
}

func printExport(val *sts.AssumeRoleOutput) {
	fmt.Printf("export AWS_ACCESS_KEY_ID=%s\n", *val.Credentials.AccessKeyId)
	fmt.Printf("export AWS_SECRET_ACCESS_KEY=%s\n", *val.Credentials.SecretAccessKey)
	fmt.Printf("export AWS_SESSION_TOKEN=%s\n", *val.Credentials.SessionToken)
}

func setEnv(val *sts.AssumeRoleOutput) {
	os.Setenv("AWS_ACCESS_KEY_ID", *val.Credentials.AccessKeyId)
	os.Setenv("AWS_SECRET_ACCESS_KEY", *val.Credentials.SecretAccessKey)
	os.Setenv("AWS_SESSION_TOKEN", *val.Credentials.SessionToken)
}

func runCommand(args []string) error {
	env := os.Environ()
	return syscall.Exec(args[0], args[1:], env)
}

func main() {
	toAssume := prepareAssumeInput()
	sess := getSession()
	role := assumeRole(sess, toAssume)

	if len(flag.Args()) > 0 {
		runCommand(flag.Args())
	} else {
		printExport(role)
	}
}
