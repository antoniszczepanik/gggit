package common

type Author struct {
	Name  string
	Email string
}

func GetAuthorFromConfig() (Author, error) {
	return Author{
		Name:  "Antoni Szczepanik",
		Email: "szczepanik.antoni@gmail.com",
	}, nil
}
