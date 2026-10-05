package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func startServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/v3/update", basicAuth(handleQuery))
	mux.HandleFunc("/nic/update", basicAuth(handleQuery))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", viper.GetInt("Server_Port")),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if viper.Get("Server_Cert") != "" && viper.Get("ServerPrivateKey") != "" {
		log.Printf("Listening (HTTPS) on %v\n", server.Addr)
		log.Fatal(server.ListenAndServeTLS(viper.GetString("Server_Cert"), viper.GetString("Server_PrivateKey")))
	} else {
		log.Printf("Listening (HTTP) on %v\n", server.Addr)
		log.Println("Warning! No certificate and key provided so serving over HTTP! You should never run DDNS_AWS in production this way! Specify certificate and private key (using config or global variables) to enable HTTPS. This will avoid the possibility of someone stealing your username/pass combination.")
		log.Fatal(server.ListenAndServe())
	}
}

func handleQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		fmt.Fprint(w, "badagent")
		return
	}

	// get required 'name' parameter
	q := r.URL.Query()
	name := q.Get("hostname")
	if name == "" {
		fmt.Fprint(w, "notfqdn")
		return
	}

	// get optional 'myip' parameter. If it's missing or not a public IPv4
	// (e.g. client behind NAT sent its WAN address), use the client's real IP
	ip := q.Get("myip")
	if parsed := net.ParseIP(ip); parsed == nil || !isPublicIPv4(parsed) {
		client, err := clientIP(r)
		if err != nil || !isPublicIPv4(client) {
			log.Printf("Rejected update for '%v': no usable public IPv4 (myip: '%v', client: '%v')\n", name, ip, client)
			fmt.Fprint(w, "dnserr")
			return
		}
		if ip != "" {
			log.Printf("Ignoring non-public myip '%v' for '%v', using client IP '%v'\n", ip, name, client)
		}
		ip = client.String()
	}

	// finally, update record
	status, err := updateRecordIP(&name, &ip)
	if err != nil {
		if status == "NOT_ACTIVE" {
			log.Printf("Error updating record: '%v' with ip:'%v'. Error was: %s: %v ", name, ip, status, err)
			fmt.Fprint(w, "nohost")
		}

		if status == "ROUTE53_UPDATE_FAIL" {
			log.Printf("Error updating record: '%v' with ip:'%v'. Error was: %s: %v ", name, ip, status, err)
			fmt.Fprint(w, "911")
		}

		return
	}

	// return proper response
	switch status {
	case "UPDATE_OK":
		log.Printf("Success: updated '%v' with ip:'%v'\n", name, ip)
		fmt.Fprintf(w, "good %v", ip)
	case "NO_CHANGE":
		log.Printf("No-change: '%v' ip is unchanged: '%v\n'", name, ip)
		fmt.Fprintf(w, "nochg %v", ip)
	default:
		log.Println("Error: 500 Internal Server Error. Unexpected success status.")
		fmt.Fprint(w, "911")
	}
}

// clientIP returns the IP of the client. X-Real-IP header is trusted only when
// the request comes from a local reverse proxy (loopback or private address).
func clientIP(r *http.Request) (net.IP, error) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return nil, err
	}
	remote := net.ParseIP(host)
	if remote == nil {
		return nil, fmt.Errorf("invalid remote address '%v'", host)
	}

	if remote.IsLoopback() || remote.IsPrivate() {
		if realIP := net.ParseIP(strings.TrimSpace(r.Header.Get("X-Real-IP"))); realIP != nil {
			return realIP, nil
		}
	}

	return remote, nil
}

// isPublicIPv4 reports whether ip is an IPv4 address routable on the internet
func isPublicIPv4(ip net.IP) bool {
	ip4 := ip.To4()
	if ip4 == nil || !ip4.IsGlobalUnicast() || ip4.IsPrivate() {
		return false
	}
	// carrier-grade NAT range 100.64.0.0/10
	if ip4[0] == 100 && ip4[1]&0xc0 == 64 {
		return false
	}
	return true
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if ok {
			userHash := sha256.Sum256([]byte(user))
			passHash := sha256.Sum256([]byte(pass))
			expectedUserHash := sha256.Sum256([]byte(viper.GetString("Server_Username")))
			expectedPassHash := sha256.Sum256([]byte(viper.GetString("Server_Password")))

			userOk := (subtle.ConstantTimeCompare(userHash[:], expectedUserHash[:]) == 1)
			passOk := (subtle.ConstantTimeCompare(passHash[:], expectedPassHash[:]) == 1)

			if userOk && passOk {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "badauth", http.StatusUnauthorized)
	})
}
