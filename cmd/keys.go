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
	"fmt"
	"io/ioutil"
	"os"

	// external packages
	log "github.com/Sirupsen/logrus"
	"github.com/jaxxstorm/go-prompt"
	"github.com/knq/ini"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	// aws
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"

	// private packages
	amazon "github.com/jaxxstorm/change-aws-credentials/pkg/aws"
)

var yes bool

// keysCmd represents the keys command
var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Rotate your AWS keys",
	Long: `Rotate your AWS secret key and secret access key, and save the new keys to 
your credentials file`,
	Run: func(cmd *cobra.Command, args []string) {

		if awsProfile == "" {
			if os.Getenv("AWS_PROFILE") == "" {
				awsProfile = "default"
			} else {
				awsProfile = os.Getenv("AWS_PROFILE")
			}
			log.Warning("Profile not specified, using default profile from credentials provider: ", awsProfile)
		}

		sess, err := amazon.New(awsProfile)

		// create an STS client and figure out who we are
		stsClient := sts.New(sess)
		currentIdentity, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})

		if err != nil {
			log.Fatal("Error getting caller identity: ", err)
		}

		// print who we are
		log.Info("Your user arn is: ", *currentIdentity.Arn)

		// create an IAM client for the current user
		iamClient := iam.New(sess)

		// get all current access keys
		currentAccessKey, err := iamClient.ListAccessKeys(&iam.ListAccessKeysInput{})

		if err != nil {
			log.Fatal("Error listing keys: ", err)
		}

		// print number of keys
		log.Info("Number of Access Keys Found: ", len(currentAccessKey.AccessKeyMetadata))

		// a slice to put keys in
		var keys []string

		// loop through all keys
		for _, key := range currentAccessKey.AccessKeyMetadata {
			// add active keys to a slice for later
			if *key.Status != "Inactive" {
				keys = append(keys, *key.AccessKeyId)
			}
			// get last used date
			lastUsed, err := iamClient.GetAccessKeyLastUsed(&iam.GetAccessKeyLastUsedInput{AccessKeyId: key.AccessKeyId})
			if err != nil {
				log.Error("Error getting last used time for key: ", key.AccessKeyId)
			}
			log.WithFields(log.Fields{"AccessKey": *key.AccessKeyId, "LastUsed": lastUsed.AccessKeyLastUsed.LastUsedDate, "Status": *key.Status}).Info("Found Access Key")
		}

		// determine the key to cycle
		var changeKey string
		if len(keys) > 1 {
			log.Info("Found multiple active keys, prompting..")
			prompt := prompt.Choose("Please specify a key to rotate", keys)
			changeKey = keys[prompt]
		} else {
			log.Info("Only one active key found, continuing..")
			// if the slice is less than 1, it'll definitely be the first in the array
			changeKey = keys[0]
		}

		if err != nil {
			log.Fatal("Error getting last used for Access Key: ", err)
		}

		// confirm operations
		var confirm bool

		if yes == false {
			confirm = prompt.Confirm("Would you like to change key: %s ? ", changeKey)
		} else {
			confirm = true
		}

		if !confirm {
			log.Fatal("Not confirmed: exiting")
		} else {

			// create the new keys
			createAccessKey, err := iamClient.CreateAccessKey(&iam.CreateAccessKeyInput{})

			if err != nil {
				log.Fatal("Error creating new Access Key: ", err)
			}

			log.Info("New Access Key: ", *createAccessKey.AccessKey.AccessKeyId)

			// deactivate the old keys
			_, err = iamClient.UpdateAccessKey(&iam.UpdateAccessKeyInput{
				AccessKeyId: aws.String(changeKey),
				Status:      aws.String("Inactive"),
			})

			if err != nil {
				log.Fatal("Error Deactivating Old Access Key: ", err)
			}

			log.Info("Old Access Key Deactivated: ", changeKey)

			home, err := homedir.Dir()

			if err != nil {
				log.Fatal("Error retrieving home directory: ", err)
			}

			credentialsPath := fmt.Sprintf("%s/.aws/credentials", home)

			data, err := ioutil.ReadFile(credentialsPath)

			if err != nil {
				log.Fatal("Error retrieving AWS credentials file: ", err)
			}

			credsFile, err := ini.LoadBytes(data)

			if err != nil {
				log.Fatal("Error loading credentials INI file: ", err)
			}

			iniProfile := credsFile.GetSection(awsProfile)

			iniProfile.SetKey("aws_access_key_id", *createAccessKey.AccessKey.AccessKeyId)
			iniProfile.SetKey("aws_secret_access_key", *createAccessKey.AccessKey.SecretAccessKey)

			log.Info("Writing new Keys to Credentials file: ", credentialsPath)

			credsFile.Write(credentialsPath)

		}

	},
}

func init() {
	RootCmd.AddCommand(keysCmd)

	keysCmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "Don't prompt for confirmation")

}
