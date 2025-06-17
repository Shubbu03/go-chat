package pkg

import (
	"net/http"
	"os"
)

func SetTokenCookies(w http.ResponseWriter, tokens *TokenPair) {
	isProduction := os.Getenv("ENV") == "production"
	domain := ""
	if isProduction {
		domain = os.Getenv("COOKIE_DOMAIN")
	}

	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    tokens.AccessToken,
		Path:     "/",
		Domain:   domain,
		MaxAge:   int(GetAccessTokenTTL().Seconds()),
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, accessCookie)

	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/auth",
		Domain:   domain,
		MaxAge:   int(GetRefreshTokenTTL().Seconds()),
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, refreshCookie)
}

func ClearTokenCookies(w http.ResponseWriter) {
	isProduction := os.Getenv("ENV") == "production"
	domain := ""
	if isProduction {
		domain = os.Getenv("COOKIE_DOMAIN")
	}

	accessCookie := &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		Domain:   domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, accessCookie)

	refreshCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/auth",
		Domain:   domain,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isProduction,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, refreshCookie)
}
