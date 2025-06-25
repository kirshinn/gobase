package main

import (
	"encoding/base64"
	"fmt"
	_ "html/template"
	"log"
	"net/http"
	"strings"
)

var users = map[string]string{
	"admin": "123",
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)

	fmt.Println("Server starting on :8087")
	log.Fatal(http.ListenAndServe(":8087", nil))
}

// Главная страница
func homeHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthorized(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	_, err := fmt.Fprintf(w, "Welcome to the protected area!")
	if err != nil {
		return
	}
}

// Страница логина
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if !isAuthorized(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	_, err := fmt.Fprintf(w, "Successfully logged in!")
	if err != nil {
		return
	}
}

// Обработчик для сброса авторизации для выхода
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Возвращаем 401 с заголовком WWW-Authenticate, чтобы сбросить кэш браузера
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	w.Header().Set("Location", "/") // Указываем, куда перенаправить
	w.WriteHeader(http.StatusUnauthorized)
	_, err := fmt.Fprintf(w, "You have been logged out. Redirecting to home...")
	if err != nil {
		return
	}
}

// Проверка авторизации
func isAuthorized(r *http.Request) bool {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}
	if !strings.HasPrefix(auth, "Basic ") {
		return false
	}
	encodedCreds := strings.TrimPrefix(auth, "Basic ")
	decodedBytes, err := base64.StdEncoding.DecodeString(encodedCreds)
	if err != nil {
		return false
	}
	creds := string(decodedBytes)
	username, password, found := strings.Cut(creds, ":")
	if !found {
		return false
	}
	storedPassword, exists := users[username]
	return exists && storedPassword == password
}
