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
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/bgentry/speakeasy"

	log "github.com/Sirupsen/logrus"
)

var cfgFile string
var awsProfile string
var userName string
var newPass string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "change-aws-credentials",
	Short: "Change your AWS credentials quickly from the cmdline",
	Long: `Allows users to quickly reset their AWS credentials without
having to burden an administrator`,
	Run: func(cmd *cobra.Command, args []string) {

		sess, err := session.NewSessionWithOptions(session.Options{
			Profile: awsProfile,
		})

		if newPass == "" {
			newPass = getPassword()
		}

		if userName == "" {
			log.Fatal("Please specify a username: See --help")
		}

		if awsProfile == "" {
			log.Warning("Profile not specified, using default from AWS_PROFILE env var: ", os.Getenv("AWS_PROFILE"))
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

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.change-aws-password.yaml)")
	RootCmd.PersistentFlags().StringVarP(&awsProfile, "awsprofile", "P", "", "AWS Profile to Change Credentials for")
	RootCmd.PersistentFlags().StringVarP(&userName, "username", "u", "", "Username to change pass for")
	RootCmd.PersistentFlags().StringVarP(&newPass, "password", "p", "", "New AWS Password for user & profile")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".change-aws-password") // name of config file (without extension)
	viper.AddConfigPath("$HOME")                // adding home directory as first search path
	viper.AutomaticEnv()                        // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debug("Using config file:", viper.ConfigFileUsed())
	}
}

func getPassword() string {
	password, _ := speakeasy.Ask("Please enter the new password: ")
	return strings.TrimSpace(password)
}
