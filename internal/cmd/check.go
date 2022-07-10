package cmd

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/go-shiori/shiori/internal/database"
	"github.com/go-shiori/shiori/internal/model"
	"github.com/spf13/cobra"
)

func checkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Find bookmarked sites that no longer exists on the internet",
		Long: "Check all bookmarks and find bookmarked sites that no longer exists on the internet. " +
			"It might take a long time depending on how many bookmarks that you have and want to check. " +
			"If there are no arguments, it will check ALL of your bookmarks.",
		Run: checkHandler,
	}

	cmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt and check ALL bookmarks")

	return cmd
}

func checkHandler(cmd *cobra.Command, args []string) {
	// Parse flags
	skipConfirm, _ := cmd.Flags().GetBool("yes")

	// If no arguments (i.e all bookmarks going to be checked), confirm to user
	if len(args) == 0 && !skipConfirm {
		confirmCheck := ""
		fmt.Print("Check ALL bookmarks? (y/N): ")
		fmt.Scanln(&confirmCheck)

		if confirmCheck != "y" {
			fmt.Println("No bookmarks checked")
			return
		}
	}

	// Convert args to ids
	ids, err := parseStrIndices(args)
	if err != nil {
		cError.Printf("Failed to parse args: %v\n", err)
		os.Exit(1)
	}

	// Fetch bookmarks from database
	filterOptions := database.GetBookmarksOptions{IDs: ids}
	bookmarks, err := db.GetBookmarks(cmd.Context(), filterOptions)
	if err != nil {
		cError.Printf("Failed to get bookmarks: %v\n", err)
		os.Exit(1)
	}

	// Create HTTP client
	httpClient := &http.Client{Timeout: time.Minute}

	// Test each bookmark item
	unreachableIDs := []int{}

	wg := sync.WaitGroup{}
	chDone := make(chan struct{})
	chProblem := make(chan int, 10)
	chMessage := make(chan interface{}, 10)
	semaphore := make(chan struct{}, 10)

	for i, book := range bookmarks {
		wg.Add(1)

		go func(i int, book model.Bookmark) {
			// Make sure to finish the WG
			defer wg.Done()

			// Register goroutine to semaphore
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
			}()

			// Ping bookmark's URL
			_, err := httpClient.Get(book.URL)
			if err != nil {
				chProblem <- book.ID
				chMessage <- fmt.Errorf("failed to reach %s: %v", book.URL, err)
				return
			}

			// Send success message
			chMessage <- fmt.Sprintf("Reached %s", book.URL)
		}(i, book)
	}

	// Watch messages from channels
	go func(nBookmark int) {
		logIndex := 0

		for {
			select {
			case <-chDone:
				cInfo.Println("Check finished")
				return
			case id := <-chProblem:
				unreachableIDs = append(unreachableIDs, id)
			case msg := <-chMessage:
				logIndex++

				switch msg.(type) {
				case error:
					cError.Printf("[%d/%d] %v\n", logIndex, nBookmark, msg)
				case string:
					cInfo.Printf("[%d/%d] %s\n", logIndex, nBookmark, msg)
				}
			}
		}
	}(len(bookmarks))

	// Wait until all download finished
	wg.Wait()
	close(chDone)

	// Print the unreachable bookmarks
	fmt.Println()

	var code int
	if len(unreachableIDs) == 0 {
		cInfo.Println("All bookmarks are reachable.")
	} else {
		sort.Ints(unreachableIDs)
		code = 1
		cError.Println("Encountered some unreachable bookmarks:")
		for _, id := range unreachableIDs {
			cError.Printf("%d ", id)
		}
		fmt.Println()
	}
	os.Exit(code)
}
