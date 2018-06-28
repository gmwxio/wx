package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/dns/v1beta2"
)

// dnsaddCmd represents the dnsadd command
var dnsaddCmd = &cobra.Command{
	Use:   "dnsadd",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: dnsaddObj.Run,
}

type dnsaddType struct {
}

var dnsaddObj = &dnsaddType{}

func (s *dnsaddType) Run(cmd *cobra.Command, args []string) {
	fmt.Printf("dnsadd\n")

	// clientID := "xxx"
	// secret := "xxx"

	// config := &oauth2.Config{
	// 	ClientID:     clientID,
	// 	ClientSecret: secret,
	// 	Endpoint:     google.Endpoint,
	// 	Scopes:       []string{dns.CloudPlatformScope, dns.CloudPlatformReadOnlyScope, dns.NdevClouddnsReadonlyScope, dns.NdevClouddnsReadwriteScope},
	// }
	// ctx := context.Background()

	// cacheFile := tokenCacheFile(config)
	// token, err := tokenFromFile(cacheFile)
	// if err != nil {
	// 	token = tokenFromWeb(ctx, config)
	// 	saveToken(cacheFile, token)
	// } else {
	// 	log.Printf("Using cached token %#v from %q", token, cacheFile)
	// }

	// oauthHttpClient := config.Client(ctx, token)

	by, err := ioutil.ReadFile("WX Cloud-80b24c5f3aa2.json")
	if err != nil {
		fmt.Printf("1. %+v\n", err)
		return
	}
	config, err := google.JWTConfigFromJSON(by,
		dns.CloudPlatformScope, dns.CloudPlatformReadOnlyScope, dns.NdevClouddnsReadonlyScope, dns.NdevClouddnsReadwriteScope)
	if err != nil {
		fmt.Printf("11. %+v\n", err)
		return
	}
	oauthHttpClient := config.Client(oauth2.NoContext)

	dnsService, err := dns.New(oauthHttpClient)
	if err != nil {
		fmt.Printf("1. %+v\n", err)
		return
	}
	mzlc := dnsService.ManagedZones.List("wx-02-cloud")
	fmt.Printf("2. %+v\n", mzlc)
	mzr, err := mzlc.Do()
	if err != nil {
		fmt.Printf("3. %+v\n", err)
		return
	}
	fmt.Printf("4. %+v\n", mzr)
	for _, mz := range mzr.ManagedZones {
		fmt.Printf("5. %+v\n", mz)
	}

}

// func saveToken(file string, token *oauth2.Token) {
// 	f, err := os.Create(file)
// 	if err != nil {
// 		log.Printf("Warning: failed to cache oauth token: %v", err)
// 		return
// 	}
// 	defer f.Close()
// 	gob.NewEncoder(f).Encode(token)
// }

// func tokenFromFile(file string) (*oauth2.Token, error) {
// 	f, err := os.Open(file)
// 	if err != nil {
// 		return nil, err
// 	}
// 	t := new(oauth2.Token)
// 	err = gob.NewDecoder(f).Decode(t)
// 	return t, err
// }

// func tokenCacheFile(config *oauth2.Config) string {
// 	hash := fnv.New32a()
// 	hash.Write([]byte(config.ClientID))
// 	hash.Write([]byte(config.ClientSecret))
// 	hash.Write([]byte(strings.Join(config.Scopes, " ")))
// 	fn := fmt.Sprintf("go-api-demo-tok%v", hash.Sum32())
// 	return filepath.Join(".", url.QueryEscape(fn))
// }

// func tokenFromWeb(ctx context.Context, config *oauth2.Config) *oauth2.Token {
// 	ch := make(chan string)
// 	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
// 	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
// 		if req.URL.Path == "/favicon.ico" {
// 			http.Error(rw, "", 404)
// 			return
// 		}
// 		if req.FormValue("state") != randState {
// 			log.Printf("State doesn't match: req = %#v", req)
// 			http.Error(rw, "", 500)
// 			return
// 		}
// 		if code := req.FormValue("code"); code != "" {
// 			fmt.Fprintf(rw, "<h1>Success</h1>Authorized.")
// 			rw.(http.Flusher).Flush()
// 			ch <- code
// 			return
// 		}
// 		log.Printf("no code")
// 		http.Error(rw, "", 500)
// 	}))
// 	defer ts.Close()

// 	config.RedirectURL = ts.URL
// 	authURL := config.AuthCodeURL(randState)
// 	go openURL(authURL)
// 	log.Printf("Authorize this app at: %s", authURL)
// 	code := <-ch
// 	log.Printf("Got code: %s", code)

// 	token, err := config.Exchange(ctx, code)
// 	if err != nil {
// 		log.Fatalf("Token exchange error: %v", err)
// 	}
// 	return token
// }

// func openURL(url string) {
// 	try := []string{"xdg-open", "google-chrome", "open"}
// 	for _, bin := range try {
// 		err := exec.Command(bin, url).Run()
// 		if err == nil {
// 			return
// 		}
// 	}
// 	log.Printf("Error opening URL in browser.")
// }

func init() {
	rootCmd.AddCommand(dnsaddCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dnsaddCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dnsaddCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
