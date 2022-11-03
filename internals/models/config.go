package models

import "time"

// Config is the config object for the application
type Config struct {
	Base struct {
		Root   string `conf:"default:Go Web Template"`
		Import string `conf:"default:import"`
	}
	Web struct {
		APIHost         string        `conf:"default:0.0.0.0:3000"`
		DebugHost       string        `conf:"default:0.0.0.0:4000"`
		ReadTimeout     time.Duration `conf:"default:5s"`
		WriteTimeout    time.Duration `conf:"default:5s"`
		ShutdownTimeout time.Duration `conf:"default:5s"`
	}
	Auth struct {
		RedirectURL    string   `conf:"default:http://localhost:3000/callback"`
		ClientID       string   `conf:"default:12345"`
		ClientSecret   string   `conf:"default:54321"`
		Scopes         []string `conf:"default:https://www.googleapis.com/auth/userinfo.email"`
		AuthURL        string   `conf:"default:https://accounts.google.com/o/oauth2/auth"`
		TokenURL       string   `conf:"default:https://oauth2.googleapis.com/token"`
		AuthStyle      int      `conf:"default:1"`
		AccessTokenURL string   `conf:"default:https://www.googleapis.com/oauth2/v2/userinfo?access_token="`
	}
	DB struct {
		Driver     string `conf:"default:postgres"`
		Connection string `conf:"default:host=db port=5432 user=postgres dbname=postgres password=postgres sslmode=disable,noprint"`
	}
}
