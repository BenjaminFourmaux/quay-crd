package quay

// <editor-fold desc="Common models"

type Avatar struct {
	Name  string `json:"name"`
	Hash  string `json:"hash"`
	Color string `json:"color"`
	Kind  string `json:"kind"`
}

// </editor-fold>

// <editor-fold desc="User models">

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

// </editor-fold>

// <editor-fold desc="Organization models>

type Organization struct {
	Name                string  `json:"name"`
	Email               string  `json:"email,omitempty"`
	Avatar              Avatar  `json:"avatar"`
	IsAdmin             bool    `json:"is_admin"`  // If the current user is Admin of this organization
	IsMember            bool    `json:"is_member"` // If the current user is a member of this organization
	Teams               []Team  `json:"teams"`
	InvoiceEmail        bool    `json:"invoice_email"`
	InvoiceEmailAddress *string `json:"invoice_email_address"`
	TagExpirationS      int     `json:"tag_expiration_s"`
	IsFreeAccount       bool    `json:"is_free_account"`
}

type CreateOrganization struct {
	Name string `json:"name"`
}

type UpdateOrganization struct {
	Email          *string `json:"email,omitempty"`
	InvoiceEmail   *bool   `json:"invoice_email,omitempty"`
	TagExpirationS *int    `json:"tag_expiration_s,omitempty"`
}

// </editor-fold>

// <editor-fold desc="Team models">

type Team struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Role        string `json:"role"` // Can be: 'admin', 'write' or 'read'
	Avatar      Avatar `json:"avatar"`
	CanView     bool   `json:"can_view"`
	RepoCount   int    `json:"repo_count"`
	MemberCount int    `json:"member_count"`
	IsSynced    bool   `json:"is_synced"`
}

type UpdateTeam struct {
	Role        string `json:"role"`        // Role of the team
	Description string `json:"description"` // Description of the team
}

type Member struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"` // Kind of the member. Can be 'user' or 'robot'
	IsRobot bool   `json:"is_robot"`
	Avatar  Avatar `json:"avatar"`
	Invited bool   `json:"invited"`
}

// </editor-fold>

// <editor-fold desc="Repository">

type Repository struct {
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	Description    string `json:"description"`
	Kind           string `json:"kind"`
	State          string `json:"state"`
	IsPublic       bool   `json:"is_public"`
	IsStarred      bool   `json:"is_starred"`
	IsOrganization bool   `json:"is_organization"`
	StatusToken    string `json:"status_token"`
	TrustEnabled   bool   `json:"trust_enabled"`
	TagExpirationS int    `json:"tag_expiration_s"`
	IsFreeAccount  bool   `json:"is_free_account"`
	CanWrite       bool   `json:"can_write"`
	CanAdmin       bool   `json:"can_admin"`
}

type CreateRepository struct {
	Repository  string `json:"repository"` // Name of the repo
	Visibility  string `json:"visibility"`
	Namespace   string `json:"namespace"`
	Description string `json:"description"`
	RepoKind    string `json:"repo_kind"`
}

type UpdateRepository struct {
	Description *string `json:"description,omitempty"`
}

type ChangeRepositoryVisibility struct {
	Visibility string `json:"visibility"`
}

// </editor-folc>
