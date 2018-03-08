package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/RadhiFadlillah/shiori/model"
	pocketApi "github.com/motemen/go-pocket/api"
	pocketAuth "github.com/motemen/go-pocket/auth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"runtime"
	"strings"
)

var (
	pocketCmd = &cobra.Command{
		Use:   "pocket",
		Short: "Manage pocket bookmarks",
	}

	//requires a consumer key to be set
	//https://getpocket.com/developer/docs/authentication
	setConsumerKeyCmd = &cobra.Command{
		Use:   "key apikey",
		Short: "Set consumer key for pocket api",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			apikey := args[0]
			viper.Set("pocket.apikey", apikey)
			err := viper.WriteConfigAs(configPath())
			if err != nil {
				cError.Println(err)
				return
			}
		},
	}

	loginPocketCmd = &cobra.Command{
		Use:   "login username",
		Short: "Login into pocket",
		Long: "Login into pocket, this command will use the consumer key set from the command shiori pocket key" +
			"and attempt to get an access key by providing the user a browser to authorize this action",
		Run: func(cmd *cobra.Command, args []string) {
			apikey := viper.Get("pocket.apikey")
			if apikey == nil {
				cError.Println("Pocket consumer key not set")
				cError.Println("Please set with: shiori pocket key")
			}
			auth, err := obtainPocketAccessToken(apikey.(string))
			if err != nil {
				cError.Println(err)
				return
			}
			viper.Set("pocket.accesstoken", auth.AccessToken)
			err = viper.WriteConfigAs(configPath())
			if err != nil {
				cError.Println(err)
				return
			}

			cTitle.Println("Login successful for " + auth.Username)
		},
	}

	importPocketCmd = &cobra.Command{
		Use:   "import",
		Short: "Import bookmarks from pocket to shiori",
		Run: func(cmd *cobra.Command, args []string) {
			offline, _ := cmd.Flags().GetBool("offline")
			all, _ := cmd.Flags().GetBool("all")
            tag, _ := cmd.Flags().GetString("tag") 


			apikey := viper.Get("pocket.apikey")
			if apikey == nil {
				cError.Println("Pocket consumer key not set")
				cError.Println("Please set with: shiori pocket key")
				return
			}
			accesstoken := viper.Get("pocket.accesstoken")
			if accesstoken == nil {
				cError.Println("Pocket access token not set")
				cError.Println("Please login to pocket: shiori pocket login")
				return
			}

			client := pocketApi.NewClient(apikey.(string), accesstoken.(string))

			//set state for retrival from pocket
			state := pocketApi.StateUnread
			if all {
				state = pocketApi.StateAll
			}

            
			options := pocketApi.RetrieveOption{
				DetailType: "complete",
				State:      state,
			}

            if tag != "" {
                options.Tag = tag
            }   

			pocketBookmarks, err := client.Retrieve(&options)

			if err != nil {
				switch err.(type) {
				case *json.UnmarshalTypeError:
					//this usually means that there we no results found
					cError.Println("Error retrieving results from pocket!\nIt is possible no",
						"results were found in your unread items, ",
						"to import all items run command with --all")
				default:
					if strings.Contains(err.Error(), "401") {
						cError.Println("Access token not valid, please relogin with shiori pocket login")
					} else {
						//unknown error
						cError.Println("Error retrieving results from pocket!")
					}
				}
				return
			}

			currentBookmarks, err := DB.GetBookmarks(false)
			if err != nil {
				cError.Println(err)
				return
			}

			bMap := make(map[string]string)
			for _, bmark := range currentBookmarks {
				url := bmark.URL
				bMap[url] = ""
			}

			var found int
			for _, pmark := range pocketBookmarks.List {
				//check if bookmark already exists if so do not try to import again
				if _, ok := bMap[pmark.ResolvedURL]; ok {
					found++
					continue
				}

				var bookmarkTitle string

				//some items added to pocket might not have a user
				//title set, in this case use the resolved title from the url
				if pmark.GivenTitle == "" {
					bookmarkTitle = pmark.ResolvedTitle
				} else {
					bookmarkTitle = pmark.GivenTitle
				}

				bookmark := model.Bookmark{
					Title:   bookmarkTitle,
					URL:     pmark.ResolvedURL,
					Excerpt: pmark.Excerpt,
				}

				bookmarkTags := []model.Tag{}

				for _, tagM := range pmark.Tags {
					newTag := model.Tag{
						Name: tagM["tag"].(string),
					}
					bookmarkTags = append(bookmarkTags, newTag)
				}

				//if pocket item has images use the first image found
				if len(pmark.Images) > 0 {
					bookmark.ImageURL = pmark.Images["1"]["src"].(string)
				}

				bookmark.Tags = bookmarkTags
				result, err := addBookmark(bookmark, offline)
				if err != nil {
					cError.Println(err)
					return
				}
				printBookmark(result)
			}
			if found > 0 {
				cTitle.Printf("%d bookmark(s) have not been imported as they already exist in shiori.\n", found)
			}
		},
	}
)

func init() {
	pocketCmd.AddCommand(loginPocketCmd)
	pocketCmd.AddCommand(setConsumerKeyCmd)

	importPocketCmd.Flags().BoolP("offline", "o", false, "Import items from pocket without fetching data from internet.")
	importPocketCmd.Flags().BoolP("all", "a", false, "Import all items from pocket, default is only to import unread items")
	importPocketCmd.Flags().StringP("tag", "t", "", "Only import items with defined tag")
	pocketCmd.AddCommand(importPocketCmd)

	rootCmd.AddCommand(pocketCmd)
}

//https://github.com/motemen/go-pocket/blob/master/cmd/pocket/main.go
func obtainPocketAccessToken(consumerKey string) (*pocketAuth.Authorization, error) {
	ch := make(chan struct{})
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if req.URL.Path == "/favicon.ico" {
				http.Error(w, "Not Found", 404)
				return
			}

			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintln(w, "Authorized.")
			ch <- struct{}{}
		}))
	defer ts.Close()
	redirectURL := ts.URL

	requestToken, err := pocketAuth.ObtainRequestToken(consumerKey, redirectURL)
	if err != nil {
		return nil, err
	}

	url := pocketAuth.GenerateAuthorizationURL(requestToken, redirectURL)
	openbrowser(url)
	<-ch

	return pocketAuth.ObtainAccessToken(consumerKey, requestToken)

}

//https://gist.github.com/hyg/9c4afcd91fe24316cbf0
func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		cError.Println(err)
		return
	}
}
