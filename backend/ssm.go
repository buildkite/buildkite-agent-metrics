package backend

import (
	"fmt"
	"os"
	"sync"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

// copied and modified from
// https://github.com/andrewoh531/gmail-attachments-to-gdrive/blob/487723750daf840e645eec1164d0b062f0e32d33/src/clients/ssm.go

var ssmClient *ssm.SSM
var once sync.Once

// Retrieve a single instance of an SSM Client
func GetSsmClient() *ssm.SSM {
	once.Do(func() {
		sess := session.Must(session.NewSession())
		ssmClient = ssm.New(sess)
	})
	return ssmClient
}

func RetrieveFromParameterStore(ssmClient ssmiface.SSMAPI, key string) string {
	output, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           &key,
		WithDecryption: aws.Bool(true),
	})

	if err != nil {
		fmt.Printf("Error retrieving SSM (%s): %v", key, err)
		os.Exit(1)
	}

	return *output.Parameter.Value
}