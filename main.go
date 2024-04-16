package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Struct untuk menyimpan data DNS Record
type DNSRecord struct {
	Name string `json:"name"`
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error loading .env file:", err)
		return
	}

	authToken := os.Getenv("AUTH_TOKEN")

	// Menangani rute "/result"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Buat klien HTTP
		client := &http.Client{}

		// Buat permintaan GET
		req, err := http.NewRequest("GET", "https://api.cloudflare.com/client/v4/zones/05aced9cfbdac55295b5e75d6c129e9f/dns_records", nil)
		if err != nil {
			http.Error(w, "Error creating request", http.StatusInternalServerError)
			return
		}

		// Atur header permintaan
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)

		// Lakukan permintaan HTTP
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Error making request", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Baca balasan
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "Error reading response body", http.StatusInternalServerError)
			return
		}

		// Parse data JSON
		var resultData map[string]interface{}
		err = json.Unmarshal(body, &resultData)
		if err != nil {
			http.Error(w, "Error parsing JSON", http.StatusInternalServerError)
			return
		}

		// Dapatkan array result
		resultArray := resultData["result"].([]interface{})

		// Persiapkan map untuk melacak setiap nama yang sudah ditampilkan
		displayedNames := make(map[string]bool)

		// Persiapkan slice untuk menyimpan nama-nama DNS record
		var records []DNSRecord

		// Loop melalui array result dan tambahkan nama-nama DNS record ke slice
		for _, item := range resultArray {
			name := item.(map[string]interface{})["name"].(string)

			// Periksa apakah nama sudah ditampilkan sebelumnya
			if !displayedNames[name] {
				record := DNSRecord{
					Name: name,
				}
				records = append(records, record)

				// Tandai nama sebagai sudah ditampilkan
				displayedNames[name] = true
			}
		}

		// Menggunakan template HTML
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			http.Error(w, "Error parsing HTML template", http.StatusInternalServerError)
			return
		}

		// Menyusun data untuk disampaikan ke template
		data := struct {
			Records []DNSRecord
		}{
			Records: records,
		}

		// Menyampaikan template dengan data
		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Error executing HTML template", http.StatusInternalServerError)
			return
		}
	})

	// Menyediakan akses ke file statis (style.css)
	staticDir := "/static/"
	http.Handle(staticDir, http.StripPrefix(staticDir, http.FileServer(http.Dir(filepath.Join(".", "templates")))))

	// Mulai server di port 8080
	fmt.Println("Server listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}
