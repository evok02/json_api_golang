package main

import (
	"os"
	jwt "github.com/golang-jwt/jwt/v4"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewApiServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/account", makeHttpHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHttpHandleFunc(s.handleAccountID)))
	router.HandleFunc("/transfer", makeHttpHandleFunc(s.handleTransfer))

	log.Print("JSON API server runnnig on address: ", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)

}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}

	if r.Method == "POST" {
		return s.handleCreateAccount(w, r)
	}

	return fmt.Errorf("usupported method %s not allowed...", r.Method)
}

func (s *APIServer) handleAccountID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccountByID(w, r)
	}

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return nil
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	WriteJSON(w, http.StatusOK, account)
	return nil
}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}

	WriteJSON(w, http.StatusOK, accounts)
	return nil
}

func withJWTAuth(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		cookie, err := getCookie(w, r, "access_token")
		if err != nil {
			fmt.Println(err)
			return
		}

		tokenStr := cookie.Value


		_, err = validateJWT(tokenStr)
		if err != nil {
			WriteJSON(w, http.StatusForbidden, ApiError{Error:"invalid token"})
			return 
		}
		

		f(w, r)
	}
}

func getCookie(w http.ResponseWriter, r *http.Request, key string) (*http.Cookie, error) {
	cookie, err := r.Cookie(key)
	if err != nil {
		return nil, err
	}

	return cookie, nil
}

func validateJWT(tokenStr string) (*jwt.Token, error) {
	secret := os.Getenv("SECRET_KEY")
	return jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	accountReq := &ReqCreateAccount{}
	err := json.NewDecoder(r.Body).Decode(accountReq)
	if err != nil {
		return fmt.Errorf("couldn't decode request body: %w", err)
	}

	account := NewAccount(accountReq.FirstName, accountReq.LastName)
	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	token, err := CreateJWT(account)
	if err != nil {
		return err
	}
	fmt.Println(token)

	setCookie(w, r, token)
	
	return WriteJSON(w, http.StatusOK, account)
}

func CreateJWT(account *Account) (string, error) {
	key := os.Getenv("SECRET_KEY")

	claims := &jwt.MapClaims{
		"expires_at": 15000,
		"account_number": account.BankNumber,
	}
	tokenStr := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return tokenStr.SignedString([]byte(key))
}

func setCookie(w http.ResponseWriter, r *http.Request, token string) {
	cookie := http.Cookie{
		Name: "access_token",
		Value: token,
		MaxAge: 3600,
		Path: "/",
		HttpOnly: false,
		Secure: true,
	}

	http.SetCookie(w, &cookie)
	fmt.Println("cookie has been sent!")
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	err = s.store.DeleteAccount(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	transferReq := &Transfer{}
	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
		return err
	}
	defer r.Body.Close()

	return WriteJSON(w, http.StatusOK, transferReq)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHttpHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func getID(r *http.Request) (int, error) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		return 0, err
	}

	return id, nil
}
