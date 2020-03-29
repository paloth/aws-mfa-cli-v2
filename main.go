package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/bigkevmcd/go-configparser"
	"github.com/manifoldco/promptui"
)

const (
	credentialFile string = "/.aws/credentials"
)

func removeBadProfile(list *configparser.ConfigParser) []string {
	return list.Sections()
}

func main() {
	var (
		home        string
		profileList []string
	)

	fmt.Println("Auth cli with mfa project")
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(2)
	}

	config, err := configparser.NewConfigParserFromFile(home + credentialFile)
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(2)
	}
	profileList = removeBadProfile(config)

	prompt := promptui.Select{
		Label: "Select profile",
		Items: profileList,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	fmt.Printf("You choose %q\n", result)

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("Enter your aws username: ")
	scanner.Scan()
	username := scanner.Text()

	fmt.Println("Enter your token: ")
	scanner.Scan()
	token := scanner.Text()

	awsSession, err := session.NewSessionWithOptions(session.Options{
		Profile: result,
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr)
		}
	}
	svcSts := sts.New(awsSession)

	identity, err := svcSts.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr)
		}
	}
	fmt.Println(identity.Account)

	tmpSession, err := svcSts.GetSessionToken(&sts.GetSessionTokenInput{
		DurationSeconds: aws.Int64(3600),
		SerialNumber:    aws.String("arn:aws:iam::" + *identity.Account + ":mfa/" + username),
		TokenCode:       aws.String(token),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr)
		}
	}
	fmt.Println(tmpSession)
}
