package slack

type Config struct {
	Ok       bool
	Error    string
	Self     User
	Channels []Channel
	Url      string
	Users    []User
}

type State struct {
	Self     User            `json:"self"`
	Users    map[string]User `json:"users"`
	Channels []Channel       `json:"channels"`
}

type User struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	RealName       string      `json:"real_name"`
	Deleted        bool        `json:"deleted"`
	Color          string      `json:"color"`
	IsAdmin        bool        `json:"is_admin"`
	IsOwner        bool        `json:"is_owner"`
	IsPrimaryOwner bool        `json:"is_primary_owner"`
	IsBot          bool        `json:"is_bot"`
	Profile        UserProfile `json:"profile"`
}

type Channel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsChannel bool   `json:"is_channel"`
	IsIM      bool   `json:"is_im"`
	// User       string
	Created    int
	Creator    string   `json:"creator"`
	IsArchived bool     `json:"is_archived"`
	IsGeneral  bool     `json:"is_general"`
	IsMember   bool     `json:"is_member"`
	Members    []string `json:"members"`
}

type UserProfile struct {
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	RealName           string `json:"real_name"`
	Title              string `json:"title"`
	RealNameNormalized string `json:"real_name_normalized"`
	Email              string `json:"email"`
}
