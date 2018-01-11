package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// New exported AWS function
func New(profile string) (*session.Session, error) {

	// grab credentials from env vars first
	// then use the config file
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{
				Profile: profile,
			},
		},
	)

	// get creds
	_, err := creds.Get()

	// create a session with creds we've used
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
	})

	if err != nil {
		return sess, err
	}

	return sess, nil

}
