package backend

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"os"
	"sync"
)

var ssmClient *ssm.SSM
var once sync.Once

func GetSsmClient() *ssm.SSM {
	once.Do(func() {
		ssmClient = ssm.New(session.Must(session.NewSession()))
	})
	return ssmClient
}

func RetrieveFromParameterStore(ssmClient *ssm.SSM, key string) string {
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
