package models

// Credentials represents stored credentials for a remote repository
type Credentials struct {
	RemoteURL string `json:"remoteUrl"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}
