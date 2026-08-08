package hme

import "strings"

func requestTypes(hostName string) (string, string) {
	if strings.Contains(hostName, "maildomainws") {
		return "application/json", "application/json"
	}
	return "application/json", "application/json, text/plain, */*"
}

func generatePayload() map[string]string {
	return map[string]string{}
}
