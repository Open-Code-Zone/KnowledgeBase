package auth

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Open-Code-Zone/cms/config"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/azureadv2"
)

func NewAuthService(store sessions.Store) {
	gothic.Store = store

	goth.UseProviders(
		azureadv2.New(
			config.Envs.AzureADClientID,
			config.Envs.AzureADClientSecret,
			buildCallbackURL("azureadv2"),
			azureadv2.ProviderOptions{
				Tenant: azureadv2.TenantType(config.Envs.AzureADTenantID),
				Scopes: []azureadv2.ScopeType{
					"User.Read",
				},
			},
		),
	)
}

func GetSessionUser(r *http.Request) (goth.User, error) {
	session, err := gothic.Store.Get(r, SessionName)
	if err != nil {
		return goth.User{}, err
	}

	u := session.Values["user"]
	if u == nil {
		return goth.User{}, fmt.Errorf("User is not authenticated! %v", u)
	}

	return u.(goth.User), nil
}

func StoreUserSession(w http.ResponseWriter, r *http.Request, u goth.User) error {
	session, _ := gothic.Store.Get(r, SessionName)

	session.Values["user"] = u
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}

func RemoveUserSession(w http.ResponseWriter, r *http.Request) {
	session, err := gothic.Store.Get(r, SessionName)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["user"] = goth.User{}
	session.Options.MaxAge = -1

	session.Save(r, w)
}

func RequireAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := GetSessionUser(r)
		if err != nil {
			log.Println("error occured getting session", err)
			http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
			return
		}

		log.Printf("user is authenticated! user: %v!", session.FirstName)
		handlerFunc(w, r)
	}
}

func buildCallbackURL(provider string) string {
	return fmt.Sprintf("%s:%s/auth/%s/callback", config.Envs.PublicHost, config.Envs.Port, provider)
}
