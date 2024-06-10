package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
)

var templates = template.Must(template.ParseFiles("templates/login.html", "templates/comics.html"))

func main() {
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/comics", handleComics)

	fmt.Println("Starting web server on :8081")
	http.ListenAndServe(":8081", nil)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		templates.ExecuteTemplate(w, "login.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		loginData := map[string]string{
			"username": username,
			"password": password,
		}
		jsonData, err := json.Marshal(loginData)
		if err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
			fmt.Println("Error encoding JSON:", err)
			return
		}

		resp, err := http.Post("http://localhost:8080/login", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			http.Error(w, "Error logging in", http.StatusInternalServerError)
			fmt.Println("Error sending POST request:", err)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading response", http.StatusInternalServerError)
			fmt.Println("Error reading response:", err)
			return
		}

		fmt.Println("Response Body:", string(body))

		if resp.StatusCode != http.StatusOK {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			fmt.Println("Invalid username or password. Response Code:", resp.StatusCode)
			return
		}

		for _, cookie := range resp.Cookies() {
			if cookie.Name == "token" {
				http.SetCookie(w, cookie)
				http.Redirect(w, r, "/comics", http.StatusSeeOther)
				return
			}
		}

		http.Error(w, "Error retrieving token", http.StatusInternalServerError)
		fmt.Println("Error retrieving token")
	}
}

func handleComics(w http.ResponseWriter, r *http.Request) {
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		query := r.URL.Query().Get("search")
		if query == "" {
			templates.ExecuteTemplate(w, "comics.html", nil)
			return
		}

		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://localhost:8080/pics?search="+url.QueryEscape(query), nil)
		if err != nil {
			http.Error(w, "Error creating request", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Authorization", "Bearer "+tokenCookie.Value)

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Error making request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading response", http.StatusInternalServerError)
			return
		}

		var comicURLs []string
		if err := json.Unmarshal(body, &comicURLs); err != nil {
			http.Error(w, "Error unmarshalling response", http.StatusInternalServerError)
			fmt.Println("Error unmarshalling response:", err)
			return
		}

		if len(comicURLs) > 3 {
			comicURLs = comicURLs[:3]
		}

		comicURLsJSON, err := json.Marshal(comicURLs)
		if err != nil {
			http.Error(w, "Error marshaling comics to JSON", http.StatusInternalServerError)
			fmt.Println("Error marshaling comics to JSON:", err)
			return
		}

		templates.ExecuteTemplate(w, "comics.html", template.JS(comicURLsJSON))
	}
}