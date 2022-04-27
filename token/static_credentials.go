package token

type staticCredentialsProvider struct {
	token *string
}

func NewStaticCredentialsProvider(token *string) Provider {
	return staticCredentialsProvider{token: token}
}

func (p staticCredentialsProvider) Get() (string, error) {
	if p.token == nil || *p.token == "" {
		return "", errNotUsable
	}
	return *p.token, nil
}

func (p staticCredentialsProvider) String() string {
	return "StaticCredentialsProvider"
}
