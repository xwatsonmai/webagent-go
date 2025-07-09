package htmlHandler

import "github.com/playwright-community/playwright-go"

func PwCookies(cookieList []playwright.Cookie) []playwright.OptionalCookie {
	var optionalCookies []playwright.OptionalCookie
	for _, cookie := range cookieList {
		optionalCookies = append(optionalCookies, playwright.OptionalCookie{
			Name:   cookie.Name,
			Value:  cookie.Value,
			Domain: playwright.String(cookie.Domain),
			Path:   playwright.String(cookie.Path),
			Expires: func() *float64 {
				if cookie.Expires != 0 {
					return playwright.Float(cookie.Expires)
				}
				return nil
			}(),
			HttpOnly: playwright.Bool(cookie.HttpOnly),
			Secure:   playwright.Bool(cookie.Secure),
			SameSite: cookie.SameSite,
		})
	}
	return optionalCookies
}
