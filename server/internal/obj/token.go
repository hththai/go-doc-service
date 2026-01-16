package obj

type AuthToken struct {
	UserId      int    `json:"userId"`
	TokenId     string `json:"tokenId"`
	AccessToken string `json:"access_token"`
}
