package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/zitadel/oidc/v2/pkg/client/rp"
	"github.com/zitadel/oidc/v2/pkg/client/rs"
	httphelper "github.com/zitadel/oidc/v2/pkg/http"
	"github.com/zitadel/oidc/v2/pkg/oidc"

	"github.com/walterwanderley/am2manager/cmd/am2server/internal/user"
	"github.com/walterwanderley/am2manager/cmd/am2server/templates"
)

var (
	clientSecret  = os.Getenv("GOOGLE_CLIENT_SECRET")
	hashKey       = "defaultkey1234567890123456789012"
	encriptionKey = []byte("defaultsessionkey123456789012345")

	sessionStore = sessions.NewCookieStore([]byte(hashKey), encriptionKey)
)

func Enabled() bool {
	return clientSecret != ""
}

func Setup(mux *http.ServeMux, db *sql.DB) error {
	redirectServer := os.Getenv("REDIRECT_SERVER")
	if redirectServer == "" {
		redirectServer = "http://localhost:8080"
	}
	hashKey = clientSecret[:19] + "am2managerkey"
	encriptionKey = append([]byte(clientSecret[:15]), []byte("am2managersession")...)
	var (
		clientID     = "60988106159-10utjruhmjssgqe3vv09oo2chnfl0j7g.apps.googleusercontent.com"
		issuer       = "https://accounts.google.com"
		callbackPath = "/auth"
		scopes       = []string{"https://www.googleapis.com/auth/userinfo.email"}
	)

	cookieHandler := httphelper.NewCookieHandler([]byte(hashKey+hashKey), encriptionKey)
	options := []rp.Option{
		rp.WithCookieHandler(cookieHandler),
		rp.WithVerifierOpts(rp.WithIssuedAtOffset(5 * time.Second)),
	}
	if clientSecret == "" {
		options = append(options, rp.WithPKCE(cookieHandler))
	}
	redirectURI := redirectServer + callbackPath

	provider, err := rp.NewRelyingPartyOIDC(issuer, clientID, clientSecret, redirectURI, scopes, options...)
	if err != nil {
		return fmt.Errorf("starting RelyingPartyOIDC: %w", err)
	}
	mux.HandleFunc("GET /login", login(provider))
	mux.HandleFunc("GET /auth/info", userInfo)
	mux.HandleFunc("GET /logout", logout)

	resourceServer, err := rs.NewResourceServerClientCredentials(issuer, clientID, clientSecret)
	if err != nil {
		return fmt.Errorf("starting ResourceServerClientCredentials: %w", err)
	}

	mux.Handle("GET "+callbackPath, rp.CodeExchangeHandler(rp.UserinfoCallback(saveUserinfo[*oidc.IDTokenClaims](resourceServer, db)), provider))

	return nil
}

func UserContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if path := r.URL.Path; path != "/liveness" && !strings.Contains(path, ".") {
			var u templates.User
			session, _ := sessionStore.Get(r, hashKey)
			if email, ok := session.Values["email"]; ok {
				u.Email = email.(string)
			}
			if name, ok := session.Values["name"]; ok {
				u.Name = name.(string)
			}
			if id, ok := session.Values["id"]; ok {
				u.ID = id.(int64)
			}
			if picture, ok := session.Values["picture"]; ok {
				u.Picture = picture.(string)
			}

			r = r.WithContext(templates.ContextWithUser(r.Context(), u))
		}

		next.ServeHTTP(w, r)
	})
}

func login(provider rp.RelyingParty) http.HandlerFunc {
	state := func() string {
		return uuid.New().String()
	}
	return rp.AuthURLHandler(state, provider)
}

func userInfo(w http.ResponseWriter, r *http.Request) {
	user := templates.UserFromContext(r.Context())
	json.NewEncoder(w).Encode(user)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, hashKey)
	delete(session.Values, "email")
	delete(session.Values, "name")
	delete(session.Values, "id")
	delete(session.Values, "picture")
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func saveUserinfo[C oidc.IDClaims](_ rs.ResourceServer, db *sql.DB) rp.CodeExchangeUserinfoCallback[C] {
	userService := user.NewService(user.New(db))
	return func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[C], state string, rp rp.RelyingParty, info *oidc.UserInfo) {
		session, _ := sessionStore.Get(r, hashKey)
		session.Options.SameSite = http.SameSiteLaxMode
		session.Options.HttpOnly = true
		session.Options.Secure = Enabled()
		session.Options.MaxAge = 86400 //1 day

		u, err := userService.GetOrInsert(r.Context(), user.UserRequest{
			Email:   info.Email,
			Name:    info.Name,
			Picture: info.Picture,
		}, db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.Values["email"] = u.Email
		session.Values["name"] = u.Name
		session.Values["id"] = u.ID
		session.Values["picture"] = u.Picture.String

		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}
