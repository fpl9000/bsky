// This is a command-line app that interacts with the BlueSky AT protocol in various ways.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

// Display usage information
func displayUsage() {
    // Get the last component of the pathname stored in os.Args[0].  Handle both Windows and
    // Unix-like pathnames.
    basename := filepath.Base(os.Args[0])

	fmt.Fprintf(os.Stderr, "Usage: %s USERNAME POSTCOUNT\n\n", basename)
	fmt.Fprintf(os.Stderr, "USERNAME   ->  BlueSky username (e.g., user.bsky.social).\n")
	fmt.Fprintf(os.Stderr, "POSTCOUNT  ->  Number of posts to fetch (a positive integer).\n")
	os.Exit(1)
}

// Fetch user posts without any authentication
func fetchUserPostsPublic(userHandle string, limit int) error {
	// Use the public API endpoint - no authentication required
	baseURL := "https://public.api.bsky.app/xrpc/app.bsky.feed.getAuthorFeed"

	params := url.Values{}
	params.Add("actor", userHandle)                  // Can use handle or DID
	params.Add("filter", "posts_and_author_threads") // Include original posts and threads
	params.Add("limit", strconv.Itoa(limit))         // Number of posts to fetch

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Simple GET request - no auth headers needed
	resp, err := http.Get(fullURL)
	if err != nil {
		return fmt.Errorf("failed to fetch posts: %v", err)
	}
	defer resp.Body.Close()

	// Parse the JSON response
	var feedResponse struct {
		Feed []struct {
			Post struct {
				URI    string `json:"uri"`
				Author struct {
					Handle      string  `json:"handle"`
					DisplayName *string `json:"displayName"`
				} `json:"author"`
				Record struct {
					Text      string `json:"text"`
					CreatedAt string `json:"createdAt"`
				} `json:"record"`
				LikeCount   *int `json:"likeCount"`
				RepostCount *int `json:"repostCount"`
				ReplyCount  *int `json:"replyCount"`
			} `json:"post"`
		} `json:"feed"`
		Cursor *string `json:"cursor"`
	}

	err = json.NewDecoder(resp.Body).Decode(&feedResponse)
	if err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	// Process the posts
	fmt.Printf("Found %d posts for user %s:\n\n", len(feedResponse.Feed), userHandle)

	for i, item := range feedResponse.Feed {
		post := item.Post
		displayName := post.Author.Handle
		if post.Author.DisplayName != nil {
			displayName = *post.Author.DisplayName
		}

		fmt.Printf("Post %d ------------------------------------------------------------\n", i+1)
		fmt.Printf("Author: %s (@%s)\n", displayName, post.Author.Handle)
		fmt.Printf("Text: %s\n", post.Record.Text)
		fmt.Printf("Created: %s\n", post.Record.CreatedAt)

		if post.LikeCount != nil {
			fmt.Printf("Likes: %d\n", *post.LikeCount)
		}
		if post.RepostCount != nil {
			fmt.Printf("Reposts: %d\n", *post.RepostCount)
		}
		if post.ReplyCount != nil {
			fmt.Printf("Replies: %d\n", *post.ReplyCount)
		}
		fmt.Println()
	}

	// Handle pagination
	if feedResponse.Cursor != nil {
		fmt.Printf("Next page cursor: %s\n", *feedResponse.Cursor)
	}

	return nil
}

func main() {
	// Check if we have exactly 2 command-line arguments
	if len(os.Args) != 3 {
		displayUsage()
	}

	username := os.Args[1]
	numPostsStr := os.Args[2]

	// Check if either argument starts with a dash
	if username[0] == '-' || numPostsStr[0] == '-' {
		displayUsage()
	}

	// Parse the number of posts
	numPosts, err := strconv.Atoi(numPostsStr)
	if err != nil || numPosts <= 0 {
		fmt.Fprintf(os.Stderr, "Error: number_of_posts must be a positive integer\n")
		displayUsage()
	}

	// Fetch posts for the specified user
	err = fetchUserPostsPublic(username, numPosts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
