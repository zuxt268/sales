package entity

import "strings"

func SplitPhone(phoneNum string) (string, string) {
	mobile := make([]string, 0)
	landline := make([]string, 0)
	for _, phone := range strings.Split(phoneNum, ",") {
		phone = strings.TrimSpace(phone)
		if phone == "" {
			continue
		}
		if strings.HasPrefix(phone, "080") || strings.HasPrefix(phone, "090") || strings.HasPrefix(phone, "070") {
			mobile = append(mobile, phone)
		} else {
			landline = append(landline, phone)
		}
	}
	mobilePhone := strings.Join(mobile, ",")
	landlinePhone := strings.Join(landline, ",")
	return mobilePhone, landlinePhone
}
