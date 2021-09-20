package options

type Redirect struct {
	RootURL        string `json:"rootURL"`
	LogoutRedirect string `json:"logoutUrl"`
}
