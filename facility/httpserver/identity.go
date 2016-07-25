package httpserver

const userIdKey = "PublicUserId"

type IdentityMap map[string]interface{}

func (im IdentityMap) PublicUserId() string {
	return im[userIdKey].(string)
}

func (im IdentityMap) SetPublicUserId(name string) {
	im[userIdKey] = name
}
