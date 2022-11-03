package main

import (
	"crypto/md5"
	"encoding/json"
	"expvar"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
	"github.com/robrohan/go-web-template/internals/handlers"
	"github.com/robrohan/go-web-template/internals/models"
	"github.com/robrohan/go-web-template/internals/repository"
	"golang.org/x/oauth2"
)

// will be replaced with git hash
var build = "develop"

var cookieName = "WB_AT"

func main() {
	if err := run(); err != nil {
		log.Println("error :", err)
		os.Exit(1)
	}
}

func run() error {
	// =========================================================================
	// Logging
	log := log.New(os.Stdout, "WB : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// =========================================================================
	// Configuration
	cfg := models.Config{}

	if err := conf.Parse(os.Args[1:], "WB", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("WB", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	var endpoint = oauth2.Endpoint{
		AuthURL:   cfg.Auth.AuthURL,
		TokenURL:  cfg.Auth.TokenURL,
		AuthStyle: oauth2.AuthStyle(cfg.Auth.AuthStyle),
	}

	oauthConfig := &oauth2.Config{
		RedirectURL:  cfg.Auth.RedirectURL,
		ClientID:     cfg.Auth.ClientID,
		ClientSecret: cfg.Auth.ClientSecret,
		Scopes:       cfg.Auth.Scopes,
		Endpoint:     endpoint,
	}

	log.Printf("%v", cfg.Auth.ClientID)

	// =========================================================================
	// App Starting
	expvar.NewString("build").Set(build)
	log.Printf("Started : Application initializing : version %q", build)
	defer log.Println("Completed")

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("Config :\n%v\n", out)

	// =========================================================================
	// Start Database
	log.Println("Initializing database support")

	db, err := repository.OpenDatabase(
		cfg.DB.Driver, cfg.DB.Connection, cfg.Base.Root)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		log.Printf("Database Stopping: %s", cfg.DB.Connection)
		db.Close()
	}()

	// =========================================================================
	// Start Debug Service
	//
	// /debug/pprof - Added to the default mux by importing the net/http/pprof package.
	// /debug/vars - Added to the default mux by importing the expvar package.
	//
	// Not concerned with shutting this down when the application is shutdown.
	log.Println("Initializing debugging support")
	go func() {
		log.Printf("Debug Listening %s", cfg.Web.DebugHost)
		log.Printf("Debug Listener closed: %v", http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux))
	}()

	// Put the API on top of the connection
	repo := repository.Attach(cfg.Base.Root, db, cfg.DB.Driver)

	// =========================================================================
	// Setup template handling
	templates := handlers.TemplateInit()

	// =========================================================================
	// Start API Service
	log.Println("Initializing API support")

	router := mux.NewRouter() // .StrictSlash(true)

	// This is just a string we send to auth provider to
	// see if they are the one sending the response
	rand.Seed(time.Now().UnixNano())
	v := rand.Int()
	randState := fmt.Sprintf("%x", v)

	env := &models.Env{
		Db:        db,
		Log:       log,
		Router:    router,
		Cfg:       &cfg,
		RandState: randState,
	}

	// Routes
	{
		router.PathPrefix("/static").Handler(
			http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

		// URLs that start with /-/ will require login
		// Note: they are still defined down below as well
		secure := router.PathPrefix("/-/").Subrouter()
		secure.Use(LoginVerify(env, repo))

		//////////////////////////
		// Non-logged in pages
		router.HandleFunc("/", handlers.ServePage(env, templates))
		router.HandleFunc("/about", handlers.ServePage(env, templates))
		//////////////////////////
		// Routes needed for auth
		router.HandleFunc("/login", handleLogin(env, oauthConfig)).Methods("GET")
		router.HandleFunc("/callback", handleCallback(env, oauthConfig, repo)).Methods("GET")
		/////////////////////////
		// Secure pages... "the app"
		secure.HandleFunc("/home", handlers.ServePage(env, templates)).Methods("GET")
	}

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      router,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	}

	log.Printf("API listening on %s", api.Addr)
	api.ListenAndServe()

	return nil
}

func handleLogin(env *models.Env, oauth *oauth2.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: randState should be unique
		url := oauth.AuthCodeURL(env.RandState)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func handleCallback(env *models.Env, oauth *oauth2.Config, repo *repository.DataRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("state") != env.RandState {
			env.Log.Printf("State is not valid")
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		token, err := oauth.Exchange(oauth2.NoContext, r.FormValue("code"))
		if err != nil {
			env.Log.Printf("Could not get token %v\n", err.Error())
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		tokenURL := env.Cfg.Auth.AccessTokenURL + token.AccessToken
		resp, err := http.Get(tokenURL)
		if err != nil {
			env.Log.Printf("Could not create access token %v\n", err.Error())
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		defer resp.Body.Close()

		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			env.Log.Printf("Could not parse response %v\n", err.Error())
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		// We have user data in content
		userInfo := models.UserInfo{}
		json.Unmarshal(content, &userInfo)

		// TODO: move this
		rand.Seed(time.Now().UnixNano())
		v := rand.Int()
		salt := fmt.Sprintf("%x", v)

		// Add user to our local database
		user := models.NewUser(userInfo.Id, userInfo.Email, userInfo.Picture)
		repo.UpsertUser(user, salt)
		user, err = repo.GetUser(userInfo.Email)
		if err != nil {
			env.Log.Printf("Could not get user: %v", err.Error())
			return
		}

		// make an entry in the users table?
		hash := md5.Sum([]byte(user.Email + user.AuthId + salt))
		addCookie(w, cookieName, fmt.Sprintf("%s:%x", user.UUID, hash), 30*24*time.Hour)

		http.Redirect(w, r, "/-/home", http.StatusFound)
	}
}

// addCookie will apply a new cookie to the response of a http request with the key/value specified.
func addCookie(w http.ResponseWriter, name, value string, ttl time.Duration) {
	expire := time.Now().Add(ttl)
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  expire,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)
}

func LoginVerify(env *models.Env, repo *repository.DataRepository) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(cookieName)
			// No cookie at all...
			if err != nil {
				env.Log.Printf("Missing auth token")
				http.Redirect(w, r, "/login", http.StatusForbidden)
				return
			}

			// Cookie, but no value
			if cookie.Value == "" {
				env.Log.Printf("Missing auth token")
				http.Redirect(w, r, "/login", http.StatusForbidden)
				return
			}

			// Cookie with value, but not the user id key
			parts := strings.Split(cookie.Value, ":")
			uuid, err := uuid.Parse(parts[0])
			if err != nil {
				env.Log.Printf("UUID malformed")
				http.Redirect(w, r, "/login", http.StatusForbidden)
				return
			}

			user, err := repo.GetUserById(uuid)
			if err != nil {
				env.Log.Printf("UUID not found")
				http.Redirect(w, r, "/login", http.StatusForbidden)
				return
			}

			// Well formatted cookie, but hash changed for some reason
			hashString := fmt.Sprintf("%s%s%s", user.Email, user.AuthId, *user.Salt)
			hash := md5.Sum([]byte(hashString))
			if fmt.Sprintf("%x", hash) != parts[1] {
				env.Log.Printf("Hashes don't match. Login might have expired")
				http.Redirect(w, r, "/login", http.StatusForbidden)
				return
			}

			// Ok, not thing wrong, move on
			next.ServeHTTP(w, r)
		})
	}
}
