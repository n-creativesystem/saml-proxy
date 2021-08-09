package saml

import (
	"encoding/base32"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
	"github.com/gorilla/securecookie"
	"github.com/n-creativesystem/saml-proxy/infra/redis"
)

const name = "saml-token"

type samlStore struct {
	con    redis.Redis
	cookie samlsp.CookieSessionProvider
}

var _ samlsp.SessionProvider = (*samlStore)(nil)

func newSessionProvider(opts samlsp.Options, con redis.Redis) samlsp.SessionProvider {
	opts.CookieName = name
	cookie := samlsp.DefaultSessionProvider(opts)
	if con == nil {
		return cookie
	}
	return &samlStore{
		con:    con,
		cookie: cookie,
	}
}

func (s *samlStore) CreateSession(w http.ResponseWriter, r *http.Request, assertion *saml.Assertion) error {
	if domain, _, err := net.SplitHostPort(s.cookie.Domain); err == nil {
		s.cookie.Domain = domain
	}
	session, err := s.cookie.Codec.New(assertion)
	if err != nil {
		return err
	}
	sessionId := strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
	value, err := s.cookie.Codec.Encode(session)
	if err != nil {
		return err
	}
	err = s.con.SetEX(r.Context(), sessionId, value, s.cookie.MaxAge).Err()
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Domain:   s.cookie.Domain,
		Value:    sessionId,
		MaxAge:   int(s.cookie.MaxAge.Seconds()),
		HttpOnly: s.cookie.HTTPOnly,
		Secure:   s.cookie.Secure || r.URL.Scheme == "https",
		SameSite: s.cookie.SameSite,
		Path:     "/",
	})
	return nil
}

func (s *samlStore) DeleteSession(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(s.cookie.Name)
	if err == http.ErrNoCookie {
		return nil
	}
	if err != nil {
		return err
	}
	err = s.con.Del(r.Context(), cookie.Value).Err()
	cookie.Value = ""
	cookie.Expires = time.Unix(1, 0) // past time as close to epoch as possible, but not zero time.Time{}
	cookie.Path = "/"
	cookie.Domain = s.cookie.Domain
	http.SetCookie(w, cookie)
	return err
}

func (s *samlStore) GetSession(r *http.Request) (samlsp.Session, error) {
	cookie, err := r.Cookie(s.cookie.Name)
	if err == http.ErrNoCookie {
		return nil, samlsp.ErrNoSession
	} else if err != nil {
		return nil, err
	}
	value, err := s.con.Get(r.Context(), cookie.Value).Result()
	if err != nil {
		return nil, samlsp.ErrNoSession
	}
	session, err := s.cookie.Codec.Decode(value)
	if err != nil {
		return nil, samlsp.ErrNoSession
	}
	return session, nil
}
