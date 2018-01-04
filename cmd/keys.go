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
	//"fmt"

	"os"

	"github.com/spf13/cobra"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/segmentio/go-prompt"

	log "github.com/Sirupsen/logrus"
)

var yes bool

// keysCmd represents the keys command
var keysCmd = &cobra.Command{
	Use:   "keys",
	Short: "Rotate your AWS keys",
	Long: `Rotate your AWS secret key and secret access key, and save the new keys to 
your credentials file`,
	Run: func(cmd *cobra.Command, args []string) {

		sess, err := session.NewSessionWithOptions(session.Options{
			Profile: awsProfile,
		})

		if err != nil {
			log.Fatal("Error creating AWS Session: ", err)
		}

		if awsProfile == "" {
			log.Warning("Profile not specified, using default from AWS_PROFILE env var: ", os.Getenv("AWS_PROFILE"))
		}

		// create an STS client and figure out who we are
		stsClient := sts.New(sess)
		currentIdentity, err := stsClient.GetCallerIdentity(&sts.GetCallerIdentityInput{})

		if err != nil {
			log.Fatal("Error getting caller identity: ", err)
		}

		log.Info("Your user arn is: ", *currentIdentity.Arn)

		// create an IAM client for the current user
		iamClient := iam.New(sess)

		// get all current access keys
		currentAccessKey, err := iamClient.ListAccessKeys(&iam.ListAccessKeysInput{})

		if err != nil {
			log.Fatal("Error listing keys: ", err)
		}

		log.Info("Number of Access Keys Found: ", len(currentAccessKey.AccessKeyMetadata))

		var keys []string

		// loop through all keys
		for _, key := range currentAccessKey.AccessKeyMetadata {
			if *key.Status != "Inactive" {
				keys = append(keys, *key.AccessKeyId)
			}
			lastUsed, err := iamClient.GetAccessKeyLastUsed(&iam.GetAccessKeyLastUsedInput{AccessKeyId: key.AccessKeyId})
			if err != nil {
				log.Error("Error getting last used time for key: ", key.AccessKeyId)
			}
			log.WithFields(log.Fields{"AccessKey": *key.AccessKeyId, "LastUsed": lastUsed.AccessKeyLastUsed.LastUsedDate, "Status": *key.Status}).Info("Found Access Key")
		}

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

		var confirm bool

		if yes == false {
			confirm = prompt.Confirm("Would you like to change key: %s ? ", changeKey)
		} else {
			confirm = true
		}

		if !confirm {
			log.Fatal("Not confirmed: exiting")
		} else {

			// create the access key
			createAccessKey, err := iamClient.CreateAccessKey(&iam.CreateAccessKeyInput{})

			if err != nil {
				log.Fatal("Error creating new Access Key: ", err)
			}

			log.Info("New Access Key: ", *createAccessKey.AccessKey.AccessKeyId)
			log.Info("New Secret Key: ", *createAccessKey.AccessKey.SecretAccessKey)
			log.Warn("Please save these to your credentials file!")

			_, err = iamClient.UpdateAccessKey(&iam.UpdateAccessKeyInput{
				AccessKeyId: aws.String(changeKey),
				Status:      aws.String("Inactive"),
			})

			if err != nil {
				log.Fatal("Error Deactivating Old Access Key: ", err)
			}

			log.Info("Old Access Key Deactivated: ", changeKey)

		}

	},
}

func init() {
	RootCmd.AddCommand(keysCmd)

	keysCmd.PersistentFlags().BoolVarP(&yes, "yes", "y", false, "Don't prompt for confirmation")

}
