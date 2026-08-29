package quay

type Avatar struct {
	Name  string `json:"name"`
	Hash  string `json:"hash"`
	Color string `json:"color"`
	Kind  string `json:"kind"`
}

type User struct {
	Anonymous           bool           `json:"anonymous"`
	Username            string         `json:"username"`
	Avatar              Avatar         `json:"avatar"`
	CanCreateRepo       bool           `json:"can_create_repo"`
	IsMe                bool           `json:"is_me"`
	Verified            bool           `json:"verified"`
	Email               string         `json:"email"`
	Logins              []string       `json:"logins"`
	InvoiceEmail        bool           `json:"invoice_email"`
	InvoiceEmailAddress *string        `json:"invoice_email_address"`
	PreferredNamespace  bool           `json:"preferred_namespace"`
	TagExpirationS      int            `json:"tag_expiration_s"`
	Prompts             []string       `json:"prompts"`
	Company             *string        `json:"company"`
	FamilyName          *string        `json:"family_name"`
	GivenName           *string        `json:"given_name"`
	Location            *string        `json:"location"`
	IsFreeAccount       bool           `json:"is_free_account"`
	HasPasswordSet      bool           `json:"has_password_set"`
	Organizations       []Organization `json:"organizations"`
	SuperUser           bool           `json:"superuser"`
}

type Organization struct {
	Name               string `json:"name"`
	Avatar             Avatar `json:"avatar"`
	CanCreateRepo      bool   `json:"can_create_repo"`
	Public             bool   `json:"public"`
	IsOrgAdmin         bool   `json:"is_org_admin"`
	PreferredNamespace bool   `json:"preferred_namespace"`
}
