package passwordreset

import "strings"

const maskCharacters = "•••"

func MaskEmail(emailAddress string) string {
	atIndex := strings.LastIndex(emailAddress, "@")
	if atIndex <= 0 {
		return maskCharacters
	}
	localPart, domain := []rune(emailAddress[:atIndex]), emailAddress[atIndex:]
	if len(localPart) <= 2 {
		return string(localPart[0]) + maskCharacters + domain
	}
	return string(localPart[0]) + maskCharacters + string(localPart[len(localPart)-1]) + domain
}
