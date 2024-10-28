package alldebrid

type GetUserDataUser struct {
	Username       string `json:"username"`
	Email          string `json:"email"`
	IsPremium      bool   `json:"isPremium"`
	IsSubscribed   bool   `json:"isSubscribed"`
	IsTrial        bool   `json:"isTrial"`
	PremiumUntil   int    `json:"premiumUntil"`
	Lang           string `json:"lang"`
	PreferedDomain string `json:"preferedDomain"`
	FidelityPoints int    `json:"fidelityPoints"`
}

type GetUserData struct {
	User GetUserDataUser `json:"user"`
}

type GetUserParams struct {
	Ctx
}

func (c APIClient) GetUser(params *GetUserParams) (APIResponse[GetUserDataUser], error) {
	response := &Response[GetUserData]{}
	res, err := c.Request("GET", "/v4/user", params, response)
	return newAPIResponse(res, response.Data.User), err
}
