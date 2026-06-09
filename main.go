package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// JSON Structs matching Instagram Data Export Format

type FollowerItem struct {
	Title          string           `json:"title"`
	MediaListData  []interface{}    `json:"media_list_data"`
	StringListData []StringListData `json:"string_list_data"`
}

type StringListData struct {
	Href      string `json:"href"`
	Value     string `json:"value"`
	Timestamp int64  `json:"timestamp"`
}

type FollowingData struct {
	RelationshipsFollowing []FollowingItem `json:"relationships_following"`
}

type FollowingItem struct {
	Title          string           `json:"title"`
	StringListData []StringListData `json:"string_list_data"`
}

type UnfollowedItem struct {
	Timestamp   int64            `json:"timestamp"`
	Media       []interface{}    `json:"media"`
	LabelValues []LabelValueItem `json:"label_values"`
	Fbid        string           `json:"fbid"`
}

type LabelValueItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Struct for serving to HTML template

type DisplayUser struct {
	Username  string `json:"username"`
	Href      string `json:"href"`
	Date      string `json:"date"`
	Timestamp int64  `json:"timestamp"`
}

type DashboardData struct {
	Unfollowers          []DisplayUser
	Fans                 []DisplayUser
	Mutuals              []DisplayUser
	RecentlyUnfollowed   []DisplayUser
	RecentFollowRequests []DisplayUser
}

var cachedData *DashboardData

func main() {
	// Initial data load
	if err := loadAllData(); err != nil {
		log.Printf("Warning during initial data load: %v", err)
	}

	// Handlers
	http.HandleFunc("/", handleDashboard)
	http.HandleFunc("/reload", handleReload)

	port := ":3030"
	fmt.Printf("Starting server at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func getLabelValue(labels []LabelValueItem, targetLabel string) string {
	for _, lv := range labels {
		if lv.Label == targetLabel {
			return lv.Value
		}
	}
	return ""
}

func formatDate(ts int64) string {
	if ts == 0 {
		return "Unknown Date"
	}
	// Format to dynamic local date
	t := time.Unix(ts, 0)
	return t.Format("02 Jan 2006, 15:04")
}

func cleanInstagramURL(username string) string {
	return "https://www.instagram.com/" + username
}

func loadAllData() error {
	var followers []DisplayUser
	var following []DisplayUser
	var unfollowed []DisplayUser

	// 1. Load followers (support followers_*.json glob)
	followerFiles, err := filepath.Glob("followers_*.json")
	if err != nil {
		return fmt.Errorf("failed to search for followers files: %v", err)
	}
	if len(followerFiles) == 0 {
		// Try root directory path if run from subdirectory
		followerFiles, _ = filepath.Glob("./followers_*.json")
	}

	for _, file := range followerFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Error reading file %s: %v", file, err)
			continue
		}

		var items []FollowerItem
		if err := json.Unmarshal(data, &items); err != nil {
			log.Printf("Error parsing JSON from %s: %v", file, err)
			continue
		}

		for _, item := range items {
			if len(item.StringListData) > 0 {
				info := item.StringListData[0]
				username := info.Value
				if username == "" {
					// Fallback to extract from href if value is empty
					parts := strings.Split(info.Href, "/")
					if len(parts) > 0 {
						username = parts[len(parts)-1]
					}
				}
				if username != "" {
					followers = append(followers, DisplayUser{
						Username:  username,
						Href:      cleanInstagramURL(username),
						Timestamp: info.Timestamp,
						Date:      formatDate(info.Timestamp),
					})
				}
			}
		}
	}

	// 2. Load following
	followingFile := "following.json"
	if _, err := os.Stat(followingFile); os.IsNotExist(err) {
		followingFile = "./following.json"
	}

	if data, err := os.ReadFile(followingFile); err == nil {
		var followingRaw FollowingData
		if err := json.Unmarshal(data, &followingRaw); err == nil {
			for _, item := range followingRaw.RelationshipsFollowing {
				username := item.Title
				var ts int64
				if len(item.StringListData) > 0 {
					ts = item.StringListData[0].Timestamp
					if username == "" {
						username = item.StringListData[0].Value
					}
				}
				if username == "" {
					// Fallback: extract username from href
					parts := strings.Split(item.StringListData[0].Href, "/")
					username = parts[len(parts)-1]
				}
				if username != "" {
					following = append(following, DisplayUser{
						Username:  username,
						Href:      cleanInstagramURL(username),
						Timestamp: ts,
						Date:      formatDate(ts),
					})
				}
			}
		} else {
			log.Printf("Error parsing following.json: %v", err)
		}
	} else {
		log.Printf("Error reading following.json: %v", err)
	}

	// 3. Load recently unfollowed
	unfollowedFile := "recently_unfollowed_profiles.json"
	if _, err := os.Stat(unfollowedFile); os.IsNotExist(err) {
		unfollowedFile = "./recently_unfollowed_profiles.json"
	}

	if data, err := os.ReadFile(unfollowedFile); err == nil {
		var unfollowedRaw []UnfollowedItem
		if err := json.Unmarshal(data, &unfollowedRaw); err == nil {
			for _, item := range unfollowedRaw {
				username := getLabelValue(item.LabelValues, "Username")
				if username == "" {
					// Fallback to Name label if username is blank
					username = getLabelValue(item.LabelValues, "Name")
				}
				if username != "" {
					unfollowed = append(unfollowed, DisplayUser{
						Username:  username,
						Href:      cleanInstagramURL(username),
						Timestamp: item.Timestamp,
						Date:      formatDate(item.Timestamp),
					})
				}
			}
		} else {
			log.Printf("Error parsing recently_unfollowed_profiles.json: %v", err)
		}
	} else {
		log.Printf("Error reading recently_unfollowed_profiles.json: %v", err)
	}

	// 4. Load recent follow requests
	var followRequests []DisplayUser
	requestsFile := "recent_follow_requests.json"
	if _, err := os.Stat(requestsFile); os.IsNotExist(err) {
		requestsFile = "./recent_follow_requests.json"
	}

	if data, err := os.ReadFile(requestsFile); err == nil {
		var requestsRaw []UnfollowedItem
		if err := json.Unmarshal(data, &requestsRaw); err == nil {
			for _, item := range requestsRaw {
				username := getLabelValue(item.LabelValues, "Username")
				if username == "" {
					username = getLabelValue(item.LabelValues, "Name")
				}
				if username != "" {
					followRequests = append(followRequests, DisplayUser{
						Username:  username,
						Href:      cleanInstagramURL(username),
						Timestamp: item.Timestamp,
						Date:      formatDate(item.Timestamp),
					})
				}
			}
		} else {
			log.Printf("Error parsing recent_follow_requests.json: %v", err)
		}
	} else {
		log.Printf("Error reading recent_follow_requests.json: %v", err)
	}

	// Compute segmentations
	followerMap := make(map[string]DisplayUser)
	for _, f := range followers {
		followerMap[f.Username] = f
	}

	followingMap := make(map[string]DisplayUser)
	for _, f := range following {
		followingMap[f.Username] = f
	}

	var unfollowersList []DisplayUser
	var mutualsList []DisplayUser
	var fansList []DisplayUser

	// Calculate Unfollowers (following but not in followers) and Mutuals
	for _, f := range following {
		if _, found := followerMap[f.Username]; !found {
			unfollowersList = append(unfollowersList, f)
		} else {
			mutualsList = append(mutualsList, f)
		}
	}

	// Calculate Fans (followers but not in following)
	for _, f := range followers {
		if _, found := followingMap[f.Username]; !found {
			fansList = append(fansList, f)
		}
	}

	// Sort lists by Timestamp descending (newest interactions first)
	sortByTimestampDesc(unfollowersList)
	sortByTimestampDesc(fansList)
	sortByTimestampDesc(mutualsList)
	sortByTimestampDesc(unfollowed)
	sortByTimestampDesc(followRequests)

	cachedData = &DashboardData{
		Unfollowers:          unfollowersList,
		Fans:                 fansList,
		Mutuals:              mutualsList,
		RecentlyUnfollowed:   unfollowed,
		RecentFollowRequests: followRequests,
	}

	log.Printf("Loaded: %d followers, %d following, %d recently unfollowed, %d follow requests. Calculated: %d unfollowers, %d fans, %d mutuals.",
		len(followers), len(following), len(unfollowed), len(followRequests), len(unfollowersList), len(fansList), len(mutualsList))

	return nil
}

func sortByTimestampDesc(list []DisplayUser) {
	sort.Slice(list, func(i, j int) bool {
		if list[i].Timestamp == list[j].Timestamp {
			return strings.ToLower(list[i].Username) < strings.ToLower(list[j].Username)
		}
		return list[i].Timestamp > list[j].Timestamp
	})
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if cachedData == nil {
		if err := loadAllData(); err != nil {
			http.Error(w, fmt.Sprintf("Error loading data: %v", err), http.StatusInternalServerError)
			return
		}
	}

	tmplFile := filepath.Join("templates", "dashboard.html")
	tmpl, err := template.ParseFiles(tmplFile)
	if err != nil {
		// Try fallback location
		tmpl, err = template.ParseFiles("./templates/dashboard.html")
		if err != nil {
			http.Error(w, fmt.Sprintf("Error loading template: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, cachedData); err != nil {
		log.Printf("Template execution error: %v", err)
	}
}

func handleReload(w http.ResponseWriter, r *http.Request) {
	if err := loadAllData(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "reloaded"})
}
