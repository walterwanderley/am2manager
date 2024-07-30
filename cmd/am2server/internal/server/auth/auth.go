package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/zitadel/oidc/v2/pkg/client/rp"
	"github.com/zitadel/oidc/v2/pkg/client/rs"
	httphelper "github.com/zitadel/oidc/v2/pkg/http"
	"github.com/zitadel/oidc/v2/pkg/oidc"

	"github.com/walterwanderley/am2manager/cmd/am2server/internal/user"
)

type contextKey int

const (
	userContext contextKey = iota
)

var (
	clientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	key          = []byte("defaultkey1234567890123456789012")
	sessionKey   = "defaultsessionkey123456789012345"

	sessionStore = sessions.NewCookieStore([]byte(sessionKey))
)

type User struct {
	ID    int64
	Email string
	Name  string
}

func Enabled() bool {
	return clientSecret != ""
}

func Setup(mux *http.ServeMux, db *sql.DB) error {
	key = []byte(clientSecret[:29] + "key")
	sessionKey = clientSecret[:25] + "session"
	var (
		clientID       = "60988106159-10utjruhmjssgqe3vv09oo2chnfl0j7g.apps.googleusercontent.com"
		clientSecret   = os.Getenv("GOOGLE_CLIENT_SECRET")
		issuer         = "https://accounts.google.com"
		redirectServer = "https://am2manager.fly.dev"
		callbackPath   = "/auth"
		scopes         = []string{"https://www.googleapis.com/auth/userinfo.email"}
	)

	cookieHandler := httphelper.NewCookieHandler(key, key)
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
		var u User
		session, _ := sessionStore.Get(r, sessionKey)
		if email, ok := session.Values["email"]; ok {
			u.Email = email.(string)
		}
		if name, ok := session.Values["name"]; ok {
			u.Name = name.(string)
		}
		if id, ok := session.Values["id"]; ok {
			u.ID = id.(int64)
		}

		r = r.WithContext(context.WithValue(r.Context(), userContext, u))

		next.ServeHTTP(w, r)
	})
}

func UserFromContext(ctx context.Context) User {
	user, _ := ctx.Value(userContext).(User)
	return user
}

func login(provider rp.RelyingParty) http.HandlerFunc {
	state := func() string {
		return uuid.New().String()
	}
	return rp.AuthURLHandler(state, provider)
}

func userInfo(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	json.NewEncoder(w).Encode(user)
}

func logout(w http.ResponseWriter, r *http.Request) {
	session, _ := sessionStore.Get(r, sessionKey)
	delete(session.Values, "email")
	delete(session.Values, "name")
	delete(session.Values, "id")
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func saveUserinfo[C oidc.IDClaims](resourceServer rs.ResourceServer, db *sql.DB) rp.CodeExchangeUserinfoCallback[C] {
	userService := user.NewService(user.New(db))
	return func(w http.ResponseWriter, r *http.Request, tokens *oidc.Tokens[C], state string, rp rp.RelyingParty, info *oidc.UserInfo) {
		session, _ := sessionStore.Get(r, sessionKey)
		session.Options.SameSite = http.SameSiteLaxMode
		session.Options.HttpOnly = true
		session.Options.Secure = Enabled()
		session.Options.MaxAge = 86400 //1 day

		u, err := userService.GetOrInsert(r.Context(), user.UserRequest{
			Email: info.Email,
			Name:  info.Name,
		}, db)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		session.Values["email"] = u.Email
		session.Values["name"] = u.Name
		session.Values["id"] = u.ID

		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
}
