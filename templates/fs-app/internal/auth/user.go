package auth

type User struct {
	ID string
}

var LocalUser User = User{
	ID: "local",
}
