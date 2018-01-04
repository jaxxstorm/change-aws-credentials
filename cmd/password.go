// Copyright © 2017 Lee Briggs <lee@leebriggs.co.uk>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"

	log "github.com/Sirupsen/logrus"
)

var userName string
var newPass string

// passwordCmd represents the password command
var passwordCmd = &cobra.Command{
	Use:   "password",
	Short: "Change your AWS Password",
	Long: `Change your AWS password using update-login-profile
without using your old password.`,
	Run: func(cmd *cobra.Command, args []string) {

		// grab credentials from env vars first
		// then use the config file
		creds := credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
			},
		)

		_, err := creds.Get()

		if err != nil {
			log.Fatal("Error getting creds")
		}

		sess, err := session.NewSession(&aws.Config{
			Credentials: creds,
		})

		if newPass == "" {
			newPass = getPassword()
		}

		if userName == "" {
			log.Fatal("Please specify a username: See --help")
		}

		if awsProfile == "" {
			log.Warning("Profile not specified, using default from credentials provider")
		}

		svc := iam.New(sess)
		input := &iam.UpdateLoginProfileInput{
			Password: aws.String(newPass),
			UserName: aws.String(userName),
		}
		_, err = svc.UpdateLoginProfile(input)
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case iam.ErrCodeEntityTemporarilyUnmodifiableException:
					log.Fatal(iam.ErrCodeEntityTemporarilyUnmodifiableException, aerr.Error())
				case iam.ErrCodeNoSuchEntityException:
					log.Fatal(iam.ErrCodeNoSuchEntityException, aerr.Error())
				case iam.ErrCodePasswordPolicyViolationException:
					log.Fatal(iam.ErrCodePasswordPolicyViolationException, aerr.Error())
				case iam.ErrCodeLimitExceededException:
					log.Fatal(iam.ErrCodeLimitExceededException, aerr.Error())
				case iam.ErrCodeServiceFailureException:
					log.Fatal(iam.ErrCodeServiceFailureException, aerr.Error())
				default:
					log.Fatal(aerr.Error())
				}
			} else {
				// Print the error, cast err to awserr.Error to get the Code and
				// Message from an error.
				log.Fatal(err.Error())
			}
			return
		}

		log.Info("Password changed successfully")

	},
}

func init() {
	RootCmd.AddCommand(passwordCmd)

	passwordCmd.PersistentFlags().StringVarP(&userName, "username", "u", "", "Username to change pass for")
	passwordCmd.PersistentFlags().StringVarP(&newPass, "password", "p", "", "New AWS Password for user & profile")

}
